import time
import os

target = "http://kv-public:80/put?key=metric&val=test"
print("Timestamp_ms,Status")

while True:
    ts = int(time.time() * 1000)
    # 500ms timeout prevents hanging on dead leader
    code = os.popen(f"curl -s -o /dev/null -w '%{{http_code}}' -m 0.5 {target}").read()
    
    if code == "200":
        print(f"{ts},UP")
    else:
        print(f"{ts},DOWN")
    
    time.sleep(0.1)