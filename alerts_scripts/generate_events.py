#!/usr/bin/env python3
"""
SIEM Event Generator for Windows brute-force rules (UTMStack).

Performs real SMB authentication attempts against target Windows hosts.
The UTMStack agent on each target captures the native Windows Security
events (4624/4625) and forwards them to the SIEM, triggering the rules
defined in rules_windows/:

  - bruteforce_attack.yml
      10+ failed 4625 for same target.user within 5 min
  - bruteforce_multiple_logon_failure_followed_by_success.yml
      10+ failed 4625 + one 4624 for same target.user within 5 min

SMB logons produce 4625/4624 with target.user populated, which is what
both rules consume.
"""
import argparse
import os
import random
import string
import sys
import time
from concurrent.futures import ThreadPoolExecutor, as_completed

try:
    import yaml
except ImportError:
    yaml = None

try:
    from impacket.smbconnection import SMBConnection, SessionError
except ImportError:
    print("ERROR: impacket is required. Install with: pip install impacket pyyaml",
          file=sys.stderr)
    sys.exit(1)


def random_password(length=14):
    chars = string.ascii_letters + string.digits + "!@#$%^&*"
    return "".join(random.choices(chars, k=length))


def smb_attempt(target, user, password, domain="", timeout=5):
    """Perform an SMB logon attempt. Returns True on success, False on auth fail."""
    try:
        conn = SMBConnection(target, target, sess_port=445, timeout=timeout)
        conn.login(user, password, domain)
        try:
            conn.logoff()
        except Exception:
            pass
        return True, "ok"
    except SessionError as e:
        return False, f"auth fail ({e.getErrorString()[0] if hasattr(e, 'getErrorString') else e})"
    except Exception as e:
        return False, f"network/proto error: {e}"


def run_bruteforce(target, user, attempts, delay, domain="", dry_run=False):
    print(f"[+] brute force  target={target} user={user!r} attempts={attempts}")
    for i in range(1, attempts + 1):
        pw = random_password()
        if dry_run:
            print(f"    [{i:>3}/{attempts}] DRY-RUN would try {user}:{pw}@{target}")
        else:
            ok, info = smb_attempt(target, user, pw, domain)
            tag = "SUCCESS" if ok else "FAIL   "
            print(f"    [{i:>3}/{attempts}] {tag} {user}:{pw}@{target}  [{info}]")
        if delay:
            time.sleep(delay)


def run_logon_success(target, user, valid_password, attempts, delay,
                      domain="", dry_run=False):
    """Multiple failures, then one valid login — triggers rule 2."""
    run_bruteforce(target, user, attempts, delay, domain, dry_run)
    print(f"[+] valid login  target={target} user={user!r}")
    if dry_run:
        print(f"    DRY-RUN would log in {user}:<valid_pass>@{target}")
        return
    ok, info = smb_attempt(target, user, valid_password, domain)
    tag = "SUCCESS" if ok else "FAIL   "
    print(f"    {tag} {user}@{target}  [{info}]")
    if not ok:
        print(f"    [!] valid creds did not authenticate — rule 2 won't fire", file=sys.stderr)


def load_config(path):
    if yaml is None:
        print("ERROR: pyyaml is required to read --config. pip install pyyaml",
              file=sys.stderr)
        sys.exit(1)
    with open(path) as f:
        return yaml.safe_load(f) or {}


def pick(cli, cfg, key, default):
    if cli is not None:
        return cli
    if key in cfg and cfg[key] is not None:
        return cfg[key]
    return default


def main():
    p = argparse.ArgumentParser(
        description="Generate Windows brute-force events for UTMStack SIEM rules",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Spray 12 bad passwords against 2 users on 1 target (triggers rule 1 & 3)
  ./generate_events.py --targets 10.0.0.50 --users alice bob

  # Trigger rule 2: failures followed by valid logon
  ./generate_events.py --targets 10.0.0.50 --users alice \\
      --scenario logon-success --valid-user alice --valid-pass 'Real-Pass!'

  # Full run with config file
  ./generate_events.py --config config.yml --scenario all
""",
    )
    p.add_argument("-c", "--config", help="YAML config file (CLI args override)")
    p.add_argument("--targets", nargs="+",
                   help="Target Windows hosts/IPs (with UTMStack agent)")
    p.add_argument("--users", nargs="+",
                   help="Usernames to spray with bad passwords")
    p.add_argument("--domain", help="Windows domain (default: workgroup/empty)")
    p.add_argument("--attempts", type=int,
                   help="Failed attempts per user (default 12, rule threshold is 10)")
    p.add_argument("--delay", type=float,
                   help="Seconds between attempts (default 1.0)")
    p.add_argument("--scenario",
                   choices=["bruteforce", "logon-success", "all"],
                   help="bruteforce=rule1, logon-success=rule2, all=both rules")
    p.add_argument("--valid-user", help="Valid user for logon-success scenario")
    p.add_argument("--valid-pass", help="Valid password for logon-success scenario")
    p.add_argument("--threads", type=int, default=1,
                   help="Parallel workers across (target,user) pairs (default 1)")
    p.add_argument("--dry-run", action="store_true",
                   help="Print what would be done without opening SMB connections")
    args = p.parse_args()

    cfg_path = args.config
    if not cfg_path:
        script_dir = os.path.dirname(os.path.abspath(__file__))
        for candidate in ("config.yml", "config.yaml", "config.example.yml"):
            full = os.path.join(script_dir, candidate)
            if os.path.isfile(full):
                cfg_path = full
                print(f"[*] auto-loaded config: {candidate}")
                break
    cfg = load_config(cfg_path) if cfg_path else {}

    targets    = args.targets    or cfg.get("targets", [])
    users      = args.users      or cfg.get("users", [])
    domain     = pick(args.domain,     cfg, "domain", "")
    attempts   = pick(args.attempts,   cfg, "attempts", 12)
    delay      = pick(args.delay,      cfg, "delay", 1.0)
    scenario   = pick(args.scenario,   cfg, "scenario", "bruteforce")
    valid_user = pick(args.valid_user, cfg, "valid_user", None)
    valid_pass = pick(args.valid_pass, cfg, "valid_pass", None)

    if not targets:
        p.error("must provide --targets (or `targets:` in config)")
    if not users:
        p.error("must provide --users (or `users:` in config)")
    if scenario in ("logon-success", "all") and (not valid_user or not valid_pass):
        p.error("--valid-user and --valid-pass are required for logon-success/all scenarios")

    # Build job list
    jobs = []
    for target in targets:
        for user in users:
            if scenario in ("bruteforce", "all"):
                jobs.append(("brute", target, user))
        if scenario in ("logon-success", "all"):
            jobs.append(("success", target, valid_user))

    print(f"[*] scenario={scenario}  targets={len(targets)}  users={len(users)}  "
          f"attempts/user={attempts}  delay={delay}s  threads={args.threads}  "
          f"dry_run={args.dry_run}")

    def worker(job):
        kind, target, user = job
        if kind == "brute":
            run_bruteforce(target, user, attempts, delay, domain, args.dry_run)
        elif kind == "success":
            run_logon_success(target, user, valid_pass, attempts, delay,
                              domain, args.dry_run)

    start = time.time()
    if args.threads > 1:
        with ThreadPoolExecutor(max_workers=args.threads) as ex:
            futures = [ex.submit(worker, j) for j in jobs]
            for f in as_completed(futures):
                exc = f.exception()
                if exc:
                    print(f"[!] worker error: {exc}", file=sys.stderr)
    else:
        for j in jobs:
            worker(j)

    elapsed = time.time() - start
    print(f"[+] done in {elapsed:.1f}s — check SIEM for triggered rules")


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\n[!] interrupted", file=sys.stderr)
        sys.exit(130)
