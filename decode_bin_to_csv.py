# decode_bin_to_csv.py
import argparse, csv, struct, datetime as dt

REC_FMT = "<Qff"
REC_SIZE = struct.calcsize(REC_FMT)

ap = argparse.ArgumentParser()
ap.add_argument("--in", dest="inp", required=True)
ap.add_argument("--out", required=True)
args = ap.parse_args()

with open(args.inp, "rb") as f, open(args.out, "w", newline="") as out:
    w = csv.writer(out)
    w.writerow(["ts_iso","epoch_ms","cpu_busy_pct","mem_used_pct"])
    while True:
        b = f.read(REC_SIZE)
        if not b: break
        ts_ns, cpu, mem = struct.unpack(REC_FMT, b)
        ms = ts_ns // 1_000_000
        w.writerow([dt.datetime.utcfromtimestamp(ms/1000).isoformat(timespec="milliseconds")+"Z", ms, f"{cpu:.2f}", f"{mem:.2f}"])