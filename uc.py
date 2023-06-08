#import undetected_chromedriver as uc
#import time
#
#chrome_options = uc.ChromeOptions()
## All arguments to hide robot automation trackers
#chrome_options.add_argument("--disable-blink-features=AutomationControlled")
#chrome_options.add_argument("--no-first-run")
#chrome_options.add_argument("--no-service-autorun")
#chrome_options.add_argument("--no-default-browser-check")
#chrome_options.add_argument("--disable-extensions")
#chrome_options.add_argument("--disable-popup-blocking")
#chrome_options.add_argument("--profile-directory=Default")
#chrome_options.add_argument("--ignore-certificate-errors")
#chrome_options.add_argument("--disable-plugins-discovery")
#chrome_options.add_argument("--incognito")
#
#driver = uc.Chrome(version_main=113, options=chrome_options, headless=True)
#driver.get("https://accounts.google.com")
#driver.save_screenshot("./s.png")
#cookies = driver.get_cookies()
#print(cookies)
#time.sleep(60)
#

import undetected_chromedriver as uc
#from selenium import webdriver as uc
import random,time,os,sys
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support    import expected_conditions as EC
import json
import sys

#chrome_options.add_argument("--user-data-dir=./profile")
#driver.delete_all_cookies()

###3
#driver.get("https://accounts.google.com/o/oauth2/v2/auth/oauthchooseaccount?redirect_uri=https%3A%2F%2Fdevelopers.google.com%2Foauthplayground&prompt=consent&response_type=code&client_id=407408718192.apps.googleusercontent.com&scope=email&access_type=offline&flowName=GeneralOAuthFlow")
#driver.get("https://accounts.google.com")
#wait = WebDriverWait(driver, 20)
#email_input = wait.until(EC.visibility_of_element_located((By.XPATH,  "//input[@id='identifierId']")))
#email_input.send_keys(GMAIL)
#driver.find_element(By.XPATH, "//div[@id='identifierNext']").click()
#password_input = wait.until(EC.visibility_of_element_located((By.XPATH,  "//input[@type='password']")))
#password_input.send_keys(PASSWORD)
#driver.find_element(By.XPATH, "//span[text()='Next']").click()
#

login = sys.argv[1]
# Login 
if login == "login":
    chrome_options = uc.ChromeOptions()
    chrome_options.add_argument("--disable-extensions")
    chrome_options.add_argument("--disable-popup-blocking")
    chrome_options.add_argument("--profile-directory=Default")
    chrome_options.add_argument("--ignore-certificate-errors")
    chrome_options.add_argument("--disable-plugins-discovery")
    chrome_options.add_argument("--incognito")
    chrome_options.add_argument("user_agent=DN")
    driver = uc.Chrome(options=chrome_options)
    driver.get("https://bard.google.com")
    #s = getpass.getpass("Press Enter after You are done login ")
    #print("Please login google bard manually...")
    wait = WebDriverWait(driver, 300000)
    work = wait.until(EC.visibility_of_element_located((By.XPATH,  "//textarea[@id='mat-input-0']")))
    cookies = driver.get_cookies()
    with open("./2.json", "w", newline='') as outputdata:
        json.dump(cookies, outputdata)
    driver.close()

# Restart session
#########################
#driver = uc.Chrome(options=chrome_options, headless=True)
chrome_options = uc.ChromeOptions()
chrome_options.add_argument("--disable-extensions")
chrome_options.add_argument("--disable-popup-blocking")
chrome_options.add_argument("--profile-directory=Default")
chrome_options.add_argument("--ignore-certificate-errors")
chrome_options.add_argument("--disable-plugins-discovery")
chrome_options.add_argument("--incognito")
#chrome_options.add_argument("--headless")
chrome_options.add_argument("user_agent=DN")
driver = uc.Chrome(options=chrome_options)

# Load cookie
driver.get("https://bard.google.com")
with open("./2.json", "r", newline='') as inputdata:
    ck = json.load(inputdata)
for c in ck:
    driver.add_cookie({k:c[k] for k in {'name', 'value'}})

# Renew with cookie
driver.get("https://bard.google.com")
wait = WebDriverWait(driver, 10)
try:
    work = wait.until(EC.visibility_of_element_located((By.XPATH,  "//textarea[@id='mat-input-0']")))
except:
    print("relogin clear 2.json")
    open("./2.json", "w").close()
    os.exit()
    




while 1:
   #time.sleep(3)
   #print("work")
   #sys.stdout.flush()

    for line in sys.stdin:
        message = line.strip()
        print("Received message:", message)
       #sys.stdout.flush()



#time.sleep(500000)
