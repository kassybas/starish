
name = "foo"

def welcome(msg):
    sh("""
        echo hi ${name}, ${msg}
    """)

welcome("hola!")
