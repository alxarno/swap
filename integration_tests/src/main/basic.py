import requests
import sys

prefix = ""


class BaseRequest:
    def __int__(self, name, path, body):
        self.name = name
        self.path = path
        self.body = body
        self.domain = prefix

    def send(self):
        try:
            r = requests.post(self.domain + self.path, data=self.body, timeout=(0.00001, 10))
            return r.json()
        except ValueError:
            print("Request :" + self.name + " failed :" + sys.exc_info()[0])
            return None


def set_prefix(_prefix):
    global prefix
    prefix = _prefix
