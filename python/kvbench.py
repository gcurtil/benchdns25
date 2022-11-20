import logging
import uuid

import pandas as pd

# https://plyvel.readthedocs.io/en/1.3.0/user.html
import plyvel

from util import Timer

logging.basicConfig(level=logging.DEBUG,
                    format='%(asctime)s %(levelname)s %(module)s %(message)s',
                    handlers=[logging.StreamHandler()])


def bench_leveldb():
    db = plyvel.DB('/tmp/testdb1/', create_if_missing=True)
    # db.put(b'key', b'value')

    val_b = uuid.uuid4().hex.encode()
    def write_entries(prefix: str, key_length: int, N: int):                    
        with db.write_batch() as wb:
            for i in range(N):            
                s1 = f"{prefix}|{i:08d}"
                k = s1.rjust(key_length, '0')
                wb.put(k.encode(), val_b)
            

    key_lengths = [8, 16, 24, 32, 48, 64, 96, 128]
    data = []
    N = 100000
    for key_length in key_lengths:
        prefix = uuid.uuid4().hex[:8]
        with Timer("write_entries, prefix: %s, N: %d", prefix, N, verbose=True) as t1:
            write_entries(prefix, key_length, N)
        data.append((prefix, key_length, N, t1.elapsed()))

    db.close()

    columns=["Prefix", "KL", "N", "dt"]
    df = pd.DataFrame(data, columns=columns)
    return df

def main():
    # parser = argparse.ArgumentParser(description='Redis perf test client')
    # parser.add_argument('--redis-host', default="localhost", help='redis host')
    # parser.add_argument('--redis-port', type=int, default=6379, help='redis port')
    # parser.add_argument('--num-runs', type=int, default=1, help='number of times to fetch data')
    # args = parser.parse_args()

    # df = bench_leveldb(args.redis_host, args.redis_port, args.num_runs)
    df = bench_leveldb()
    #with pd.option_context('display.float_format', '{:0.3f}'.format):
    with pd.option_context('display.precision', 3):
        logging.info("bench_leveldb timings: \n%s", df)

if __name__ == '__main__':
    main()
