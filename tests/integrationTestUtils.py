import sys

def count_running_tests(expectedNumber):
    """
    Read all input from stdin (until EOF). Check that it contains the string "kssh passed test N" for N from 1 to
    expectedNumber (inclusive on both sides). Exit with an error if it does not find all of the passed tests.
    :param expectedNumber:  The highest expect test that should have passed
    :return:                None. Calls exit() in all scenarios.
    """
    all_stdin = sys.stdin.read()
    for i in range(1, expectedNumber + 1):
        test_str = "kssh passed test %s" % i
        if test_str not in all_stdin:
            print("Did not find '%s' in logs: Missing test success!" % test_str)
            exit(42)
    if "kssh passed test %s" % (expectedNumber + 1) in all_stdin:
        print('kssh logs report passing test #%s, did you remember to update integrationTest.sh to increment the test count?')
        exit(42)
    exit(0)

if __name__ == "__main__":
    subcommand = sys.argv[1]
    if subcommand == "count":
        count_running_tests(int(sys.argv[2]))
    else:
        # Error on a bogus subcommand so we fail fast
        exit(1)