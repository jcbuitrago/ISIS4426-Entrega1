# ANB Rising Stars — Platform & Architecture Overview

## 1. Purpose
ANB Rising Stars is a basketball talent showcase. Players:
- Register, manage profile & avatar.
- Upload short raw skill videos.
- Videos are processed asynchronously (trim, normalize, brand, thumbnail).
- Vote (max 2 votes per user) on processed videos.
- View rankings and leaderboards.

Core states: `uploaded → processing → processed` (or `failed`).

## 2. High-Level Features
- Auth (JWT): signup/login/profile ([back/app/routers/auth_handler.go](../../back/app/routers/auth_handler.go))
- Protected video upload + personal list ([back/app/routers/video_handler.go](../../back/app/routers/video_handler.go))
- Public gallery of processed videos ([back/app/routers/public_handler.go](../../back/app/routers/public_handler.go))
- Voting (2 per user) + rankings
- Job status tracking ([`async.SQSEnqueuer.GetStatus`](../../back/app/async/sqs_client.go))
- S3-based storage (uploads vs processed) ([back/internal/s3client/s3client.go](../../back/internal/s3client/s3client.go))
- Asynchronous FFmpeg pipeline (worker) ([back/cmd/worker/main.go](../../back/cmd/worker/main.go))
- Resilient status updates in PostgreSQL ([db/init.sql](../../db/init.sql))
- CORS-configured API for S3-hosted SPA ([back/main.go](../../back/main.go))

## 3. Architecture (Production)
ASCII layout (private compute, public entry via ALB):

```
Browser (SPA on S3)
        |
        v
Application Load Balancer (HTTP/HTTPS, public)
        |
        v
API Auto Scaling Group (min=1 desired=1 max=3, private subnets)
        | \
        |  \--> RDS PostgreSQL (metadata, users, votes, job_status)
        |       (Security Group allows 5432 from API + Worker SGs)
        |
        +----> S3 (uploads bucket: private originals)
        +----> S3 (processed bucket: public processed assets)
        +----> SSM Parameter Store (bucket names, secrets)
        +----> SQS (queue for video processing jobs)

Worker Auto Scaling Group (min=1 desired=1 max=3, private subnets)
        |--> SQS (consume messages)
        |--> S3 (download original / upload processed)
        |--> RDS (update status, URLs)
        |--> SSM (config)
```

Key considerations:
- Both ASGs in private subnets → need NAT or VPC endpoints for S3 + SSM + SQS.
- API is never directly exposed—only through ALB.
- S3 processed bucket objects are served directly (public read or via CloudFront if added).

## 4. Data Model (Simplified)
Tables: `users`, `videos`, `votes`, `job_status` ([db/init.sql](../../db/init.sql)).
Video fields: origin S3 key (`origin_url`), processed file URL (`processed_url`), thumbnail URL (`thumb_url`), status, votes.

Model definition: [`models.Video`](../../back/app/models/video.go)

## 5. Upload & Processing Flow
1. Client sends multipart upload (`title`, `video_file`) → [`routers.VideosHandler.Create`](../../back/app/routers/video_handler.go).
2. File stored directly in S3 uploads bucket via [`s3client.UploadToUploads`](../../back/internal/s3client/s3client.go).
3. DB record created via [`services.VideoService.Create`](../../back/app/services/video_service.go) → [`repos.VideoRepoPG.Create`](../../back/app/repos/video_repo_pg.go).
4. Job enqueued on SQS: [`async.SQSEnqueuer.EnqueueVideoProcessing`](../../back/app/async/sqs_client.go). Initial job status inserted in `job_status`.
5. Worker polls SQS (external process / consumer code not shown here—extendable) and runs pipeline:
   - Download original (`DownloadFromUploads`)
   - Trim (`trimTo30`) → scale (`to720p16x9`) → intro/outro concat (`concatIntroMainOutro`) → mute audio → thumbnail (`extractThumbnail`). See helpers: [`trimTo30`](../../back/cmd/worker/proc_video.go), [`to720p16x9`](../../back/cmd/worker/proc_video.go), [`concatIntroMainOutro`](../../back/cmd/worker/proc_video.go), [`extractThumbnail`](../../back/cmd/worker/proc_video.go).
   - Upload processed + thumbnail (`UploadToProcessed`).
   - Update DB URLs: [`services.VideoService.UpdateProcessedURL`](../../back/app/services/video_service.go), [`services.VideoService.UpdateThumbURL`](../../back/app/services/video_service.go).
   - Final status set to `processed`: [`services.VideoService.UpdateStatus`](../../back/app/services/video_service.go).
6. Frontend polls job status (`GET /api/jobs/{id}`) via [`routers.JobsHandler.GetJobStatus`](../../back/app/routers/jobs_handlers.go) (file path not listed; adapt if present).
7. Public listing (`GET /api/public/videos`) only returns processed entries: [`routers.PublicHandler.ListVideos`](../../back/app/routers/public_handler.go).

## 6. SQS Usage
- Enqueue: [`async.SQSEnqueuer.EnqueueVideoProcessing`](../../back/app/async/sqs_client.go) creates structured JSON payload.
- Attributes include `JobType`, `VideoID` etc. to allow filtering or future DLQ routing.
- Health check: [`async.SQSEnqueuer.Ping`](../../back/app/async/sqs_client.go) fetches queue attributes (used in `/api/readyz` in [`main`](../../back/main.go)).
- Job status lifecycle stored in PostgreSQL via [`async.SQSEnqueuer.createJobStatus`](../../back/app/async/sqs_client.go) and updates with [`async.SQSEnqueuer.SetStatus`](../../back/app/async/sqs_client.go).

Why SQS:
- Decouples API scale from processing throughput.
- Eliminates need for shared Redis in private network.
- Native durability and visibility timeouts support retries (future extension).

## 7. CORS & SPA Integration
- Allowed origins configured in [`main`](../../back/main.go) (list: local dev + `FRONTEND_URL`).
- Must not use `*` together with credentials (`Authorization` header). Explicit origins enforced.
- SPA base URL set via `VITE_API_BASE_URL` (see frontend calls: [`front/src/api.js`](../../front/src/api.js)).

## 8. Security & Access
- Private subnets isolate API & Worker.
- Security Groups:
  - ALB → API (8080)
  - API + Worker → RDS (5432)
  - Outbound to S3/SSM/SQS via NAT or interface endpoints.
- SSM used for bucket names: [`s3client.NewFromSSM`](../../back/internal/s3client/s3client.go) reads `/anb/s3/uploads-bucket` and `/anb/s3/processed-bucket`.
- JWT middleware guards protected routes (see auth paths in [`main`](../../back/main.go)).

## 9. Error Handling & Resilience
- Each pipeline stage sets granular status codes in job_status via [`async.SQSEnqueuer.SetStatus`](../../back/app/async/sqs_client.go) (e.g. `failed:trim`, `failed:upload_processed`) making frontend feedback precise.
- Failures leave original upload intact; processed artifacts only appear on success.
- Thumbnail extracted early; if later steps fail no public listing will show the video.

## 10. Voting & Rankings
- Vote endpoints: `/api/public/videos/{id}/vote`, `/api/public/videos/{id}/vote` (DELETE) in [`routers.PublicHandler.Vote`](../../back/app/routers/public_handler.go) & `Unvote`.
- Remaining votes for user: [`routers.PublicHandler.MyVotes`](../../back/app/routers/public_handler.go).
- Rankings endpoint `/api/public/rankings` (file not expanded—see same handler file).
- DB ensures one row per (video_id,user_id); votes column denormalized for fast ordering.

## 11. Key Code Entry Points
- Initialization: [`main.main`](../../back/main.go)
- S3 client: [`s3client.NewFromSSM`](../../back/internal/s3client/s3client.go)
- Enqueue job: [`async.SQSEnqueuer.EnqueueVideoProcessing`](../../back/app/async/sqs_client.go)
- Update status: [`services.VideoService.UpdateStatus`](../../back/app/services/video_service.go)
- Upload handler: [`routers.VideosHandler.Create`](../../back/app/routers/video_handler.go)
- Public listing: [`routers.PublicHandler.ListVideos`](../../back/app/routers/public_handler.go)

## 12. Operational Endpoints
- Health: `/api/healthz` (static OK) ([back/main.go](../../back/main.go))
- Ready: `/api/readyz` (DB + SQS ping) ([back/main.go](../../back/main.go))
- Job status: `/api/jobs/{id}` (DB lookup) (jobs handler file)
- Processed videos: `/api/public/videos` (ranked by votes desc)

## 13. Scaling Considerations
- API ASG (min=1 desired=1 max=3) handles spikes in auth/list traffic.
- Worker ASG (min=1 desired=1 max=3) can scale independently based on queue depth or CPU.
- SQS buffers during bursts—no backpressure on upload route.
- FFmpeg temporary work dir in `/tmp` cleaned per job to avoid disk bloat.

## 14. Future Enhancements (Concise)
- DLQ for failed SQS messages.
- Presigned uploads direct from browser (bypass API binary upload).
- CloudFront distribution for processed bucket.
- Metrics around per-stage processing time.
- Idempotent reprocessing endpoint.

## 15. Quick Trace (Example)
Upload request → [`VideosHandler.Create`](../../back/app/routers/video_handler.go) → S3 (uploads) → DB row (status=uploaded) → SQS message → Worker consumes → FFmpeg steps (helpers in [`proc_video.go`](../../back/cmd/worker/proc_video.go)) → Upload processed + thumb → DB update (URLs + status=processed) → Appears in `/api/public/videos`.

## 16. Testing References
- Service tests: [`video_service` tests](../../back/app/services/video_services_test.go) validate status & URL updates.
- Repos (PostgreSQL queries): [`VideoRepoPG`](../../back/app/repos/video_repo_pg.go).

## 17. Constraints
- Max upload size: 100MB (see constant in [`routers.VideosHandler.Create`](../../back/app/routers/video_handler.go)).
- Max trimmed duration: 30s (`trimTo30`).
- Aspect normalization: 1280x720 with padding (`to720p16x9`).
- Two votes per user enforced in logic of [`PublicHandler.MyVotes`](../../back/app/routers/public_handler.go).

## 18. Environment Variables (Core)
- `DB_DSN` (PostgreSQL)
- `SQS_QUEUE_URL`
- `AWS_REGION`
- `FRONTEND_URL` (CORS origin)
- SSM parameters: `/anb/s3/uploads-bucket`, `/anb/s3/processed-bucket`

## 19. Summary
The system cleanly separates synchronous user interactions (upload/auth/list) from CPU-intensive media processing via SQS + Worker ASG, uses S3 for durable media storage, and maintains transparent status tracking in PostgreSQL. Minimal coupling enables independent scaling and clear operational monitoring.