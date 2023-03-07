# 简介

这是一个有趣的小 demo，想测试是否可以通过 GPT-3.5 为众多语法繁杂的测绘搜索引擎建立一个 `巴别塔`，来实现通用的自然语言测绘引擎，打通自然语言到测绘语法的最后一公里。

项目基于 project-discovery 的 [uncover](https://github.com/projectdiscovery/uncover) 改造而成，加入了自然语言到查询语法的翻译。

注意：由 GPT-3.5 生成的测绘引擎语法常出现语法错误

目前支持的引擎：
- FOFA
- 360 Quake
- Censys

## 怎么玩
1. 将 OpenAI Token 设置为环境变量 `OPENAI_KEY`
2. 按照官方指导正常配置 `uncover`
3. 使用 -fofa/-quake/-censys 传入自然语言，改造后的 `uncover` 会使用 GPT-3.5-turbo 尽可能的将输入翻译成指定测绘引擎的语法：

## 一些好玩的示例

### 美国所有开放 3306 端口的主机

```
$ env OPENAI_KEY=YOUR_KEY_HERE ./uncover-turbo -v -fofa '搜索美国开放了3306端口的主机' -json -delay 5 -r -l 10

  __  ______  _________ _   _____  _____
 / / / / __ \/ ___/ __ \ | / / _ \/ ___/
/ /_/ / / / / /__/ /_/ / |/ /  __/ /    
\__,_/_/ /_/\___/\____/|___/\___/_/ v1.0.2

                projectdiscovery.io

[DBG] Translate to fofa query: "port="3306" && country="US""
[fofa] {"timestamp":1678181098,"source":"fofa","ip":"193.227.114.8","port":3306,"host":"g-b.cn"}
[fofa] {"timestamp":1678181098,"source":"fofa","ip":"193.227.114.8","port":3306,"host":"ejjq.com"}
[fofa] {"timestamp":1678181098,"source":"fofa","ip":"193.227.114.8","port":3306,"host":"tttuuu.com"}
[fofa] {"timestamp":1678181098,"source":"fofa","ip":"193.227.114.8","port":3306,"host":"shangye.biz"}
[fofa] {"timestamp":1678181098,"source":"fofa","ip":"193.227.114.8","port":3306,"host":"yyyccc.com"}
[fofa] {"timestamp":1678181098,"source":"fofa","ip":"193.227.114.8","port":3306,"host":"rencai.biz"}
[fofa] {"timestamp":1678181098,"source":"fofa","ip":"193.227.114.8","port":3306,"host":"kkkggg.com"}
[fofa] {"timestamp":1678181098,"source":"fofa","ip":"193.227.114.8","port":3306,"host":"gongqiu.biz"}
[fofa] {"timestamp":1678181098,"source":"fofa","ip":"193.227.114.8","port":3306,"host":"dinggou.biz"}
[fofa] {"timestamp":1678181098,"source":"fofa","ip":"193.227.114.8","port":3306,"host":"3-1.cn"}
[fofa] {"timestamp":1678181098,"source":"fofa","ip":"193.227.114.8","port":3306,"host":"fuwu.biz"}
[fofa] {"timestamp":1678181098,"source":"fofa","ip":"47.89.255.228","port":3306,"host":"ptb2bvip.com"}
```

### 搜索翻墙机场面板

```
$ env OPENAI_KEY=YOUR_KEY_HERE ./uncover-turbo -v -fofa '搜索翻墙机场面板' -json -delay 5 -r -l 10

  __  ______  _________ _   _____  _____
 / / / / __ \/ ___/ __ \ | / / _ \/ ___/
/ /_/ / / / / /__/ /_/ / |/ /  __/ /    
\__,_/_/ /_/\___/\____/|___/\___/_/ v1.0.2

                projectdiscovery.io

[DBG] Translate to fofa query: "title="翻墙机场" && (body="ssr-panel" || body="v2board" || body="naiveproxy" || body="vpnpanel" || body="soga" || body="trojan-panel")"
[fofa] {"timestamp":1678181237,"source":"fofa","ip":"172.67.176.139","port":443,"host":"翻墙机场.net"}
[fofa] {"timestamp":1678181237,"source":"fofa","ip":"202.182.108.34","port":443,"host":"clashios.com"}
[fofa] {"timestamp":1678181237,"source":"fofa","ip":"45.32.85.17","port":443,"host":"clashnode.xyz"}
[fofa] {"timestamp":1678181237,"source":"fofa","ip":"172.67.147.149","port":443,"host":"sub-gfwairport.download"}
[fofa] {"timestamp":1678181237,"source":"fofa","ip":"172.67.147.149","port":80,"host":"sub-gfwairport.download"}
[fofa] {"timestamp":1678181237,"source":"fofa","ip":"172.67.180.72","port":443,"host":"翻墙机场.xyz"}
[fofa] {"timestamp":1678181237,"source":"fofa","ip":"104.21.19.241","port":443,"host":"gfwairport.icu"}
```

### 搜索所有没有鉴权的 redis

```
$ env OPENAI_KEY=YOUR_KEY_HERE ./uncover-turbo -v -fofa '搜索所有没有鉴权的 redis' -json -delay 5 -r -l 10

  __  ______  _________ _   _____  _____
 / / / / __ \/ ___/ __ \ | / / _ \/ ___/
/ /_/ / / / / /__/ /_/ / |/ /  __/ /    
\__,_/_/ /_/\___/\____/|___/\___/_/ v1.0.2

                projectdiscovery.io

[DBG] Translate to fofa query: "port="6379" && body="*-NOAUTH*""
[fofa] {"timestamp":1678181323,"source":"fofa","ip":"61.xx.36.131","port":6379,"host":""}
[fofa] {"timestamp":1678181323,"source":"fofa","ip":"114.xx.46.124","port":6379,"host":""}
[fofa] {"timestamp":1678181323,"source":"fofa","ip":"114.x.16.120","port":6379,"host":""}
[fofa] {"timestamp":1678181323,"source":"fofa","ip":"114.xx.74.43","port":6379,"host":""}
[fofa] {"timestamp":1678181323,"source":"fofa","ip":"114.xx.85.92","port":6379,"host":""}
[fofa] {"timestamp":1678181323,"source":"fofa","ip":"114.xx.54.191","port":6379,"host":""}
[fofa] {"timestamp":1678181323,"source":"fofa","ip":"114.xx.15.113","port":6379,"host":""}
[fofa] {"timestamp":1678181323,"source":"fofa","ip":"114.xx.31.93","port":6379,"host":""}
[fofa] {"timestamp":1678181323,"source":"fofa","ip":"114.xx.18.106","port":6379,"host":""}
[fofa] {"timestamp":1678181323,"source":"fofa","ip":"114.xx.73.158","port":6379,"host":""}
```

### 搜索所有没有鉴权的 elasticsearch

```
$ env OPENAI_KEY=YOUR_KEY_HERE ./uncover-turbo -v -fofa '搜索所有没有鉴权的elasticsearch' -json -delay 5 -r -l 10

  __  ______  _________ _   _____  _____
 / / / / __ \/ ___/ __ \ | / / _ \/ ___/
/ /_/ / / / / /__/ /_/ / |/ /  __/ /    
\__,_/_/ /_/\___/\____/|___/\___/_/ v1.0.2

                projectdiscovery.io

[DBG] Translate to fofa query: "body="You Know, for Search" && status_code!="401""
[fofa] {"timestamp":1678181738,"source":"fofa","ip":"8.xxx.46.19","port":8084,"host":""}
[fofa] {"timestamp":1678181738,"source":"fofa","ip":"47.x.49.11","port":87,"host":""}
[fofa] {"timestamp":1678181738,"source":"fofa","ip":"116.x.129.215","port":8003,"host":""}
[fofa] {"timestamp":1678181738,"source":"fofa","ip":"44.x.97.208","port":9200,"host":"cribl.cloud"}
[fofa] {"timestamp":1678181738,"source":"fofa","ip":"35.x.49.186","port":9200,"host":""}
[fofa] {"timestamp":1678181738,"source":"fofa","ip":"34.x.210.219","port":9200,"host":""}
[fofa] {"timestamp":1678181738,"source":"fofa","ip":"47.x.218.16","port":8004,"host":""}
[fofa] {"timestamp":1678181738,"source":"fofa","ip":"35.x.157.44","port":9200,"host":"cribl.cloud"}
```