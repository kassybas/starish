
# load("loaded.star", "loaded")

def welcome(msg="foo", prefix="PREPRE", shout=False):
    print(prefix, msg)
    if shout:
        print("SHOUTING!!!!")
    # THISS DOESNOT WORK
    sh("echo this is shell")

# loaded.trial("okay")

print("---init")

hello = "okay"


print("okay")


#
#
# starish welcome "hello"