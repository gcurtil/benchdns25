import logging
import time

import pandas as pd


class Timer:
    def __init__(self, msg, *msg_args, verbose=False):
        self.start = 0.0
        self.end = 0.0
        self.msg = msg
        self.msg_args = msg_args
        self.verbose = verbose

    def __enter__(self):
        #logging.info("Timer enter %s", self)
        self.start = time.perf_counter()
        return self

    def __exit__(self, exc_type, exc_value, exc_tb):
        self.end = time.perf_counter()
        #logging.info("Timer exit %s", self)
        if self.verbose:
            ms = self.elapsed() * 1000.0
            logging.info(f'{self.msg} - %f ms', *self.msg_args, ms)

    def elapsed(self):
        return (self.end - self.start)


class TableTimer:
    class Timer:
        def __init__(self, tbl_timer, key):
            self.tbl_timer = tbl_timer
            self.key = key
            self.t0 = self.t1 = 0

        def __enter__(self):
            # self.start = time.perf_counter()
            # self.start_mono = time.monotonic()
            self.t0 = self.new_timing()
            return self

        def __exit__(self, exc_type, exc_value, exc_tb):
            #self.end = time.perf_counter()            
            #self.end_mono = time.monotonic()            
            # self.tbl_timer.timings.append( (
            #     self.key, self.end - self.start, self.end_mono - self.start_mono
            # ))
            # logging.info("Timer exit, keys %s, %f", self.key, self.elapsed())
            self.t1 = self.new_timing()
            self.tbl_timer.timings.append( (
                self.key, 
                self.t1[0] - self.t0[0], self.t1[1] - self.t0[1],
                self.t1[2] - self.t0[2], self.t1[3] - self.t0[3],
            ))
            #logging.info("Timer exit, keys %s, %f", self.key, self.elapsed())

        def new_timing(self):
            return time.perf_counter(), time.perf_counter_ns(), time.monotonic(), time.monotonic_ns()

        def elapsed(self):
            return (self.t1[0] - self.t0[0])

    def __init__(self):
        self.timings = []

    def timer(self, key):
        return TableTimer.Timer(self, key)
    
    def table(self):
        df = pd.DataFrame(
            self.timings,
            columns=["key", "elapsed", "elapsed_ns", "elapsed_mono", "elapsed_mono_ns"]
        )
        return df
