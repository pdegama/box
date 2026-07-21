# Box SMTP
Box is scalable SMTP server! 💻✨
It handles sending & receiving emails effortlessly and integrates RabbitMQ for smooth, asynchronous processing. Ideal for handling bulk emails in modern applications.

## 🔧 Features:
- SMTP server/client
- RabbitMQ support
- Built for scalability

## Concept
Box SMTP mainly work in SMTP Server Wrapper, It make one server or more than one server (two, three, etc...), Any server recive email and it will be store into database or add into queqe and send/forward to main centrel server.

## TODO
- Enhanced status codes in smtp
- send mail to multiple rcpts
- `exp=` support in  in spf

## Config File

### Config file location
- `/etc/box.yaml`
- `/etc/box.yml`
- `./config.yaml`
- `./config.yml`

### Config file structure
```yaml
name: Box - SMTP
hostname: localhost
max_clients: 5

port: 25

tls:
  starttls: true
  key: key.pem
  cert: cert.pem

spf_check: true
message_size: 1024000
max_recipients: -1
check_mailbox: false

amqp_conf:
  host: localhost
  port: 5672
  username:
  password:
  queue: box-receiver-1

client:
  hostname: localhost
  worker: 2
  log_dir: ./log
  log_file: email.log

  amqp_conf:
    host: localhost
    port: 5672
    username:
    password:
    queue: box-sender-1
    status_queue: box-status-1

log_dir: ./log
log_file: email.log
dev: false
```

## How work
### Receive mail
mail contine in `receiver` queue. mail is store in yaml format, config receiver queue using config file
```yaml
uid: 1723707191VlKoL1I2_1
time: 2006-01-02T15:04:05Z07:00
success: true
cmds: 6/6
tls: true
ptr_ip:
    server_ptr: mx.myworkspacel.ink
    server_ip: 2a01:4f9:c012:5d00::1
    client_ptr: rellit.email
    client_ip: 2a01:4f9:c012:d77b::1
domain: rellit.email
ptr_match: true
spf_fail: false
spf_status: PASS
from: hello@parthka.dev
recipients:
    - my@myworkspacel.ink
use_bdat: false
data: |+
    DKIM-Signature: v=1; a=rsa-sha256; c=simple/simple; d=parthka.dev;
    	s=default; t=1723707190;
    	bh=EVfAHeUMDygbJe0SkMWJHjgXGjtiTLZnMQbyWqzsrCY=;
    	h=Date:To:From:Subject;
    	b=elWNDcJ9XMttrWbCwun4GW0WHGgL/L/DH+reRS0PmCFjfyE/vTNiAkx1rZjwRexJ2
    	 ibyb2GY3ShZzrrr8XH0EWa/qZVaQZkXPoZ7ilPtzLYBEs7W1TaXlnrAu/w31RFBnzG
    	 /AHhDSmWKww4ay53BHK1ZI7YKAr8B8ZPI30lNNWKnMc/0aTj0MuS037pAOnPIaNfy6
    	 7Vy1NS4KuOCLHUuutFrSJ9pCb6Xe+KGnhhfZAorHZWlhpRk0CRwJ7S4FrKxoiFuwn+
    	 jlfgum4qlO/i508w+2dAw+7Zsdcev++kcIOI9yk7Z8PuoTmkRTguic0aeKDzFNpq1+
    	 +YRT9rTu6SukQ==
    Received: from [IPV6:2402:a00:162:3778:ed5d:c9bd:60c:6b25] (unknown [IPv6:2402:a00:162:3778:ed5d:c9bd:60c:6b25])
    	by rellit.email (Postfix) with ESMTPSA id 9DEE240CEA
    	for <my@myworkspacel.ink>; Thu, 15 Aug 2024 07:33:10 +0000 (UTC)
    Message-ID: <474f3010-779f-4900-8daf-3dce5cda34ce@parthka.dev>
    Date: Thu, 15 Aug 2024 13:03:07 +0530
    MIME-Version: 1.0
    User-Agent: Mozilla Thunderbird
    Content-Language: en-US
    To: my@myworkspacel.ink
    From: Parth <hello@parthka.dev>
    Subject: Hello , From Box!
    Content-Type: text/plain; charset=UTF-8; format=flowed
    Content-Transfer-Encoding: 7bit

    Hello!

```

### Send mail
for send, mail yaml format publish in `sender` queue, config sender queue using config file
```yaml
uid: 1723707191VlKoL1I2
from: hello@box.tld
recipient: hello@idk.btw
data: |+
    DKIM-Signature: v=1; a=rsa-sha256; c=simple/simple; d=parthka.dev;
    	s=default; t=1723707190;
    	bh=EVfAHeUMDygbJe0SkMWJHjgXGjtiTLZnMQbyWqzsrCY=;
    	h=Date:To:From:Subject;
    	b=elWNDcJ9XMttrWbCwun4GW0WHGgL/L/DH+reRS0PmCFjfyE/vTNiAkx1rZjwRexJ2
    	 ibyb2GY3ShZzrrr8XH0EWa/qZVaQZkXPoZ7ilPtzLYBEs7W1TaXlnrAu/w31RFBnzG
    	 /AHhDSmWKww4ay53BHK1ZI7YKAr8B8ZPI30lNNWKnMc/0aTj0MuS037pAOnPIaNfy6
    	 7Vy1NS4KuOCLHUuutFrSJ9pCb6Xe+KGnhhfZAorHZWlhpRk0CRwJ7S4FrKxoiFuwn+
    	 jlfgum4qlO/i508w+2dAw+7Zsdcev++kcIOI9yk7Z8PuoTmkRTguic0aeKDzFNpq1+
    	 +YRT9rTu6SukQ==
    Received: from [IPV6:2402:a00:162:3778:ed5d:c9bd:60c:6b25] (unknown [IPv6:2402:a00:162:3778:ed5d:c9bd:60c:6b25])
    	by rellit.email (Postfix) with ESMTPSA id 9DEE240CEA
    	for <my@myworkspacel.ink>; Thu, 15 Aug 2024 07:33:10 +0000 (UTC)
    Message-ID: <474f3010-779f-4900-8daf-3dce5cda34ce@parthka.dev>
    Date: Thu, 15 Aug 2024 13:03:07 +0530
    MIME-Version: 1.0
    User-Agent: Mozilla Thunderbird
    Content-Language: en-US
    To: my@myworkspacel.ink
    From: Parth <hello@parthka.dev>
    Subject: Hello , From Box!
    Content-Type: text/plain; charset=UTF-8; format=flowed
    Content-Transfer-Encoding: 7bit

    Hello!

```

## Mail Status
when client send mail than client put status yaml format in `status` queue, config status queue using config file, it is contine success(bool), status, errors, etc...
```yaml
time: 2006-01-02T15:04:05Z07:00
uid: 1723707191VlKoL1I2
success: true
status: SUCCESS
errors: []
temperror: false
anyclienterror: false
```
When error
```yaml
time: 2006-01-02T15:04:05Z07:00
uid: 1723707191VlKoL1I2
success: false
status: FAIL
errors:
    - domain: mx.myworkspacel.ink.
      error: |-
        Delivery Error: 550 - email doesn't delivered because sender
        domain [cockatielone.bi] does not
        designate 2a01:4f9:c012:5d00::1 as
        permitted sender.
      code: 550
      servererror: false
    - domain: mx.myworkspacel.ink.
      error: |-
        Delivery Error: 550 - email doesn't delivered because sender
        domain [cockatielone.bi] does not
        designate 65.21.57.194 as
        permitted sender.
      code: 550
      servererror: false
temperror: false
anyclienterror: false
```

Status:
- `SUCCESS` - deliver successful
- `TRYAGAIN` - any temp error like 4yz, timeout, etc... - send mail after few minute (if same TRYAGAIN status return many time send email than stop send and redirect to fail)
- `FAIL` - fail to deliver mail, dont send send mail again
