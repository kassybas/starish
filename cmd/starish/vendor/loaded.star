
def trial(msg):
    print("This works", msg)


config = {
    "name": "myname",
    "age": 21
}

loaded = module("loaded", trial=trial, config=config)