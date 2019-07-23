import os
import sys

def count_running_tests(expected_number):
    """
    Read all input from stdin (until EOF). Check that it contains the string "kssh passed test N" for N from 1 to
    expectedNumber (inclusive on both sides). Exit with an error if it does not find all of the passed tests.
    :param expectedNumber:  The highest expect test that should have passed
    :return:                None. Calls exit() in all scenarios.
    """
    all_stdin = sys.stdin.read()
    for i in range(1, expected_number + 1):
        test_str = "kssh passed test %s" % i
        if test_str not in all_stdin:
            print("Did not find '%s' in logs: Missing test success!" % test_str, file=sys.stderr)
            exit(42)
    if "kssh passed test %s" % (expected_number + 1) in all_stdin:
        print('kssh logs report passing test #%s, did you remember to update integrationTest.sh to increment the test count?' % (expected_number + 1), file=sys.stderr)
        exit(42)
    exit(0)

def check_logs(expected_number, expected_principals):
    username = os.environ.get('KEYBASE_USERNAME', None)
    if username is None:
        print('Failed to get kssh username from the environment!')
        exit(42)
    cnt = 0
    all_stdin = sys.stdin.read()
    for line in all_stdin.splitlines():
        if "Processing SignatureRequest from user=%s" % username in line and "principals:%s, expiration:+1h, pubkey:ssh-ed25519" % expected_principals in line:
            cnt += 1
    if cnt != expected_number:
        print('found %s audit logs, expected %s' % (cnt, expected_number))
        exit(42)
    exit(0)

if __name__ == "__main__":
    subcommand = sys.argv[1]
    if subcommand == "count":
        count_running_tests(int(sys.argv[2]))
    if subcommand == "logcheck":
        check_logs(int(sys.argv[2]), sys.argv[3])
    else:
        # Error on a bogus subcommand so we fail fast
        exit(1)