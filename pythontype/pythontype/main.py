# pythontype - Python implementation of zootype typing test

import argparse

VERSION = "dev"

def main():
    parser = argparse.ArgumentParser(description="Terminal-based typing test")
    parser.add_argument("-v", "--version", action="version", version=f"pythontype {VERSION}")
    parser.parse_args()
    
    print("Hello from pythontype")
