
load("vendor/loaded.star", "loaded")

def foo(msg="foo", prefix="PREPRE", shout=False):
    print(prefix, msg)
    if shout:
        print("SHOUTING!!!!")
    sh("echo this is shell")

loaded.trial("okay")