
name = "foo"
booly = True
nilly = None
tupy = (1,2,3,4,5)
funky = {
    "hey": "you",
    "staring":"back"
    # "and": 
}
listy = [1,2,"4","6"]

def welcome(msg):
    setty= set([1,2,3,4,5])
    sh("""
        echo hi ${name}, ${msg}
        env
    """, shield_env=True)

welcome("hola!")
