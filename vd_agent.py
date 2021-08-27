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
        print('-'*20)
        print(config)
        print('-'*20)

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
        print('running single iteration')
        config = self.config
        try:
            for set in config['sets']:
                print(f'starting set {set["name"]}')
                set_opts = {
                    **self.config['global_options'],
                    **set['options']
                }
                try:
                    with yt_dlp.YoutubeDL(set_opts) as ytdlp:
                        print('clearing yt-dlp cache')
                        print('-'*20)
                        ytdlp.cache.remove()
                        print('-'*20)
                        print('starting yt-dlp')
                        print('-'*20)
                        ytdlp.download([set['url']])
                        print('-'*20)
                        print('set complete!')
                        print('\n'*2, end=None)
                except Exception as e:
                    oops(e, f'error running ytdlp.download(): {e}')
            print('iteration complete!')
            print('\n'*2, end=None)
            print('jobs info:')
            print('-'*20)
            self.sched.print_jobs()
            print('-'*20)
            print('\n'*2, end=None)
        except Exception as e:
            oops(e, f'error running single iteration: {e}')

    def __signal_handler(self):
        print('shutting down')
        sys.exit()


def oops(e, err_string):
    print(err_string)
    traceback.print_exc()
    raise e


# class YTDLPLogger():
#     def debug(self, msg):
#         print(msg)

#     def warning(self, msg):
#         print(msg)

#     def error(self, msg):
#         print(msg)


if __name__ == '__main__':
    agent = VideoDownloadAgent()
    try:
        print('running agent')
        agent.run()
    except Exception as e:
        oops(e, f'error while running agent: {e}')
