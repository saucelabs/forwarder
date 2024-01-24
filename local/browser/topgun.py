import threading
import random
import time

from playwright.sync_api import sync_playwright
from playwright._impl._api_types import TimeoutError

headless = False
concurrent = 3
load_timeout = 7000
monkey_testing_timeout = 3000

# List of URLs to open.
urls = [
    "https://www.alibaba.com/",
    "https://www.aliexpress.com/",
    "https://www.amazon.com/",
    "https://www.bbc.com/",
    "https://www.cnn.com/",
    "https://www.ebay.com/",
    "https://www.foxnews.com",
    "https://www.nytimes.com/",
    "https://www.theguardian.com/",
    "https://www.washingtonpost.com/",
    "https://www.wsj.com/"
]

# Talkative websites never achieve networkidle state, instead wait for talkative_wait seconds.
talkative = set([
    "https://www.nytimes.com/",
    "https://www.washingtonpost.com/"
])
talkative_wait = 5

firefox_user_prefs = {
    "network.proxy.backup.ssl": "localhost",
    "network.proxy.backup.ssl_port": 3128,
    "network.proxy.http": "localhost",
    "network.proxy.http_port": 3128,
    "network.proxy.share_proxy_settings": True,
    "network.proxy.ssl": "localhost",
    "network.proxy.ssl_port": 3128,
    "network.proxy.type": 1,

    "security.OCSP.enabled": 0,
}


def open_page():
    with sync_playwright() as p:
        browser = p.firefox.launch(headless=headless, firefox_user_prefs=firefox_user_prefs)
        context = browser.new_context()

        perm = list(range(len(urls)))
        random.shuffle(perm)

        for i in perm:
            url = urls[i]

            # Open the URL
            print(f'{threading.current_thread().name} opening {url}')

            page = context.new_page()
            page.goto(url)

            # Wait for the page to load.
            try:
                if url in talkative:
                    page.wait_for_load_state("domcontentloaded", timeout=load_timeout)
                    time.sleep(talkative_wait)
                else:
                    page.wait_for_load_state("networkidle", timeout=load_timeout)
            except TimeoutError:
                print(f'{threading.current_thread().name} timed out waiting for {url}')

            # Monkey testing
            page.add_script_tag(url="https://cdnjs.cloudflare.com/ajax/libs/gremlins.js/0.1.0/gremlins.min.js")
            page.add_script_tag(content="""gremlins.createHorde().unleash();""")
            time.sleep(monkey_testing_timeout/1000)

            page.close()

        context.close()


def main():
    threads = []
    for i in range(concurrent):
        t = threading.Thread(target=open_page, name=f'Thread-{i+1}')
        t.start()
        threads.append(t)
    for thread in threads:
        thread.join()


if __name__ == "__main__":
    main()
