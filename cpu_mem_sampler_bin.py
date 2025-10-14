# cpu_mem_sampler_bin.py
import argparse, os, socket, struct, time
import psutil

REC_FMT = "<Qff"  # epoch_ns (uint64), cpu_busy_pct (float32), mem_used_pct (float32)
REC_SIZE = struct.calcsize(REC_FMT)

def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--out", required=True, help="output binary file (write to /run for tmpfs)")
    ap.add_argument("--interval-ms", type=int, default=10)
    ap.add_argument("--duration", type=float, default=60)
    ap.add_argument("--flush-every", type=int, default=500)  # buffer bursts
    args = ap.parse_args()

    interval = max(args.interval_ms, 1) / 1000.0
    psutil.cpu_times_percent(interval=0.0)  # prime

    os.makedirs(os.path.dirname(args.out), exist_ok=True)
    buf = bytearray(REC_SIZE * args.flush_every)
    idx = 0
    start = time.time()
    with open(args.out, "wb", buffering=0) as f:  # unbuffered; we’ll batch in buf
        while True:
            cpu = 100.0 - psutil.cpu_times_percent(interval=interval).idle
            mem = psutil.virtual_memory().percent
            ts_ns = time.time_ns()
            struct.pack_into(REC_FMT, buf, (idx % args.flush_every) * REC_SIZE, ts_ns, float(cpu), float(mem))
            idx += 1
            if idx % args.flush_every == 0:
                f.write(buf)
            if args.duration > 0 and (time.time() - start) >= args.duration:
                # flush remainder
                rem = (idx % args.flush_every)
                if rem:
                    f.write(memoryview(buf)[:rem * REC_SIZE])
                break

if __name__ == "__main__":
    main()