from datetime import datetime, timedelta
import sys
import signal
import time
import traceback
import yt_dlp
import yaml
from apscheduler.schedulers.background import BackgroundScheduler


class VideoDownloadAgent():
    __ONE_DAY_IN_SECONDS = 60 * 60 * 24

    def run(self):
        print('loading configuration file')
        try:
            with open("/app/config.yaml") as config_file:
                config = yaml.safe_load(config_file)
                self.config = config
        except Exception as e:
            oops(e, f'error opening configuration file: {e}')

        print('printing supplied configuration...')
        print('ALL OPTIONS:')
        print_block(config)

        sched = BackgroundScheduler()
        self.sched = sched
        try:
            print('adding main job')
            sched.start()
            # add the job and tell it to start 1s from now
            sched.add_job(self.__run_single_iteration, 'interval',
                          seconds=config['interval'], next_run_time=datetime.now() + timedelta(seconds=1))
            sched.print_jobs()
        except Exception as e:
            oops(e, f'error starting job scheduler: {e}')

        signals_to_listen_for = [signal.SIGHUP, signal.SIGINT]
        for sig in signals_to_listen_for:
            signal.signal(sig, self.__signal_handler)

        print('waiting for a shutdown signal...')
        while True:
            time.sleep(self.__ONE_DAY_IN_SECONDS)

    def __run_single_iteration(self):
        print_block('running single iteration', 1)
        config = self.config
        try:
            for set in config['sets']:
                print_block(f'starting set {set["name"]}', 1)
                set_opts = {
                    'logger': YTDLPLogger(),
                    **self.config['global_options'],
                    **set['options']
                }
                try:
                    print('creating yt_dlp instance')
                    print_divider()
                    with yt_dlp.YoutubeDL(set_opts) as ytdlp:
                        print_divider()
                        print_empty_lines(1)

                        print('clearing yt-dlp cache')
                        print_divider()
                        ytdlp.cache.remove()
                        print_divider()
                        print_empty_lines(1)

                        print('starting yt-dlp download')
                        print_divider()
                        ytdlp.download([set['url']])
                        print_divider()
                        print_empty_lines(2)
                except Exception as e:
                    oops(e, f'error running ytdlp.download(): {e}')
            print_block('iteration complete! awaiting next interval...', 2)
            print_block('jobs info:')
            self.sched.print_jobs()
            print_divider()
            print_empty_lines(2)
        except Exception as e:
            oops(e, f'error running single iteration: {e}')

    def __signal_handler(self):
        print('shutting down')
        sys.exit()


def oops(e, err_string):
    print(err_string)
    traceback.print_exc()
    raise e


def print_block(s, n=0):
    print_divider()
    print(s)
    print_divider()
    if n > 0:
        print_empty_lines(n)


def print_empty_lines(n=2):
    print('\n'*n, end=None)


def print_divider():
    print('-'*50)


class YTDLPLogger():
    def info(self, msg):
        print(f'\t{msg}')

    def debug(self, msg):
        print(f'\t{msg}')

    def warning(self, msg):
        print(f'\t{msg}')

    def error(self, msg):
        print(f'\t{msg}')


if __name__ == '__main__':
    agent = VideoDownloadAgent()
    try:
        print_block('running agent')
        agent.run()
    except Exception as e:
        oops(e, f'error while running agent: {e}')
