#!/bin/python

import os
import sys
import threading
import json
import os.path
import logging
import queue
import subprocess
import multiprocessing


CONFIG_FILE = os.path.join(os.path.dirname(__file__), 'config.json')
VERMAKE_PATH = ''
VERSION_CONFIG = ''
UPLOAD_CONFIG = ''
SOURCE_DIR = ''

upload_queue = queue.Queue()

def parse(version_number, date, dest_dir):
    pass

def do_pkg(version_number, date, dest_dir):
    ths = []
    for task in parse(version_number, date, dest_dir):
        ths.append(threading.Thread(target=vermake, args=(task,)))
    for th in ths:
        th.start()
    for th in ths:
        th.join()

def do_upload(queue):
    ths = []
    while True:
        item = queue.get()
        th = threading.Thread(target=upload, args=(item,))
        ths.append(th)
        th.start()
    for th in ths:
        th.join()

def parse_upload():
    pass

def upload(item):
    pass

def main():
    pass

def vermake(args, queue):
    subprocess.run([VERMAKE_PATH].extend(args), check=True)
    queue.put(args)


if __name__ == "__main__":
    pass