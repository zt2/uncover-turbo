# 简介

这是一个有趣的小 demo，想测试是否可以通过 GPT-3.5 为众多语法繁杂的测绘搜索引擎建立一个 `巴别塔`，来实现通用的自然语言测绘引擎，打通自然语言到测绘语法的最后一公里。

项目基于 project-discovery 的 [uncover](https://github.com/projectdiscovery/uncover) 改造而成，加入了自然语言到查询语法的翻译。

注意：由 GPT-3.5 生成的测绘引擎语法常出现语法错误

## 怎么玩
1. 将 OpenAI Token 设置为环境变量 `OPENAI_KEY`
2. 按照官方指导正常配置 `uncover`
3. 接下来就可以开始玩了，传入自然语言作为查询语法，改造后的 `uncover` 会使用 GPT-3.5-turbo 尽可能的将输入翻译成 `shodan`, `fofa`, `quake`, `zoomeye` 等测绘引擎的语法

### 查询美国开放了3306端口的主机

```
$ env OPENAI_KEY=YOUR-KEY ./uncover-turbo -v -q '查询美国开放了3306端口的主机' -json

  __  ______  _________ _   _____  _____
 / / / / __ \/ ___/ __ \ | / / _ \/ ___/
/ /_/ / / / / /__/ /_/ / |/ /  __/ /    
\__,_/_/ /_/\___/\____/|___/\___/_/ v1.0.2

                projectdiscovery.io

[DBG] Translate to fofa query: port="3306" && country="US"
[DBG] Translate to quake query: find host where port=3306 and country="US"
[DBG] Translate to zoomeye query: port:3306 country:"United States"
[DBG] Translate to shodan query: country:"US" port:"3306" has_ipv4:true has_ipv6:true
[zoomeye] unexpected status code 403 received from https://api.zoomeye.org/host/search?query=find host where port=3306 and country="US"&page=1
[quake] json: cannot unmarshal string into Go struct field Response.code of type int
[quake] json: cannot unmarshal string into Go struct field Response.code of type int
[quake] json: cannot unmarshal string into Go struct field Response.code of type int
[quake] {"timestamp":1678132746,"source":"quake","ip":"34.28.129.216","port":3306,"host":"216.129.28.34.bc.googleusercontent.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"198.154.193.40","port":3306,"host":"198-154-193-40.unifiedlayer.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"13.56.100.193","port":3306,"host":"ec2-13-56-100-193.us-west-1.compute.amazonaws.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"144.126.248.35","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"5.8.41.232","port":3306,"host":"kvm11.ny2.gcorelabs.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"67.225.147.52","port":3306,"host":"asphost214.asphostserver.org"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"198.154.193.194","port":3306,"host":"web.5lottomonkeys.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"3.138.46.229","port":3306,"host":"ec2-3-138-46-229.us-east-2.compute.amazonaws.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"34.192.89.16","port":3306,"host":"ec2-34-192-89-16.compute-1.amazonaws.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"13.66.29.25","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"198.154.193.213","port":3306,"host":"198-154-193-213.unifiedlayer.com"}
[zoomeye] {"timestamp":1678132756,"source":"zoomeye","ip":"160.72.106.98","port":1443,"host":""}
[zoomeye] {"timestamp":1678132756,"source":"zoomeye","ip":"80.84.30.24","port":5060,"host":""}
[zoomeye] {"timestamp":1678132756,"source":"zoomeye","ip":"160.72.82.98","port":1443,"host":""}
[zoomeye] {"timestamp":1678132756,"source":"zoomeye","ip":"68.195.208.146","port":32791,"host":""}
[zoomeye] {"timestamp":1678132756,"source":"zoomeye","ip":"142.252.127.72","port":8478,"host":""}
[zoomeye] {"timestamp":1678132756,"source":"zoomeye","ip":"68.195.251.14","port":32791,"host":""}
[zoomeye] {"timestamp":1678132756,"source":"zoomeye","ip":"68.195.113.210","port":32791,"host":""}
[zoomeye] {"timestamp":1678132756,"source":"zoomeye","ip":"68.195.251.30","port":32791,"host":""}
[zoomeye] {"timestamp":1678132756,"source":"zoomeye","ip":"68.197.133.90","port":32791,"host":""}
[zoomeye] {"timestamp":1678132756,"source":"zoomeye","ip":"68.199.0.118","port":32791,"host":""}
[zoomeye] {"timestamp":1678132756,"source":"zoomeye","ip":"68.193.153.214","port":32791,"host":""}
[zoomeye] {"timestamp":1678132756,"source":"zoomeye","ip":"68.206.15.78","port":32791,"host":""}
[zoomeye] {"timestamp":1678132756,"source":"zoomeye","ip":"68.207.60.23","port":32791,"host":""}
[zoomeye] {"timestamp":1678132756,"source":"zoomeye","ip":"142.252.127.84","port":8478,"host":""}
[zoomeye] {"timestamp":1678132756,"source":"zoomeye","ip":"68.193.1.133","port":32791,"host":""}
[zoomeye] {"timestamp":1678132756,"source":"zoomeye","ip":"68.195.194.134","port":32791,"host":""}
[zoomeye] {"timestamp":1678132756,"source":"zoomeye","ip":"68.195.222.218","port":32791,"host":""}
[zoomeye] {"timestamp":1678132756,"source":"zoomeye","ip":"68.195.225.146","port":32791,"host":""}
[zoomeye] {"timestamp":1678132756,"source":"zoomeye","ip":"68.197.43.83","port":32791,"host":""}
[zoomeye] {"timestamp":1678132756,"source":"zoomeye","ip":"68.194.250.72","port":32791,"host":""}
```

### 查询使用 CF 作为 CDN 的主机

```
$ env OPENAI_KEY=YOUR-KEY ./uncover-turbo -v -q '使用cloudflare cdn的主机' -json -delay 5 -r

  __  ______  _________ _   _____  _____
 / / / / __ \/ ___/ __ \ | / / _ \/ ___/
/ /_/ / / / / /__/ /_/ / |/ /  __/ /    
\__,_/_/ /_/\___/\____/|___/\___/_/ v1.0.2

                projectdiscovery.io

[DBG] Translate to fofa query: "http.title=\"Attention Required! | Cloudflare\""
[DBG] Translate to quake query: "cloudflare cdn 主机" -> "cdns('cloudflare')"
[DBG] Translate to zoomeye query: http.component:cloudflare && http.component:server
[DBG] Translate to shodan query: `http.html: "Server: cloudflare"`
[quake] json: cannot unmarshal string into Go struct field Response.code of type int
[quake] json: cannot unmarshal string into Go struct field Response.code of type int
[quake] json: cannot unmarshal string into Go struct field Response.code of type int
[zoomeye] {"timestamp":1678133907,"source":"zoomeye","ip":"144.24.14.233","port":443,"host":""}
[zoomeye] {"timestamp":1678133907,"source":"zoomeye","ip":"141.148.178.242","port":443,"host":""}
[zoomeye] {"timestamp":1678133907,"source":"zoomeye","ip":"141.148.169.88","port":443,"host":""}
[zoomeye] {"timestamp":1678133907,"source":"zoomeye","ip":"141.148.181.1","port":443,"host":""}
[zoomeye] {"timestamp":1678133907,"source":"zoomeye","ip":"141.148.148.13","port":443,"host":""}
[zoomeye] {"timestamp":1678133907,"source":"zoomeye","ip":"144.126.213.82","port":80,"host":""}
[zoomeye] {"timestamp":1678133907,"source":"zoomeye","ip":"138.197.234.171","port":443,"host":""}
[zoomeye] {"timestamp":1678133907,"source":"zoomeye","ip":"18.216.88.52","port":9000,"host":""}
[zoomeye] {"timestamp":1678133907,"source":"zoomeye","ip":"8.28.167.10","port":18080,"host":""}
[zoomeye] {"timestamp":1678133907,"source":"zoomeye","ip":"212.188.136.39","port":8080,"host":""}
[zoomeye] unexpected status code 403 received from https://api.zoomeye.org/host/search?query="cloudflare cdn 主机" -> "cdns('cloudflare')"&page=2
[zoomeye] {"timestamp":1678133913,"source":"zoomeye","ip":"13.214.152.247","port":34599,"host":""}
[zoomeye] {"timestamp":1678133913,"source":"zoomeye","ip":"182.160.16.234","port":10035,"host":""}
[zoomeye] {"timestamp":1678133913,"source":"zoomeye","ip":"8.213.129.15","port":7707,"host":""}
[zoomeye] {"timestamp":1678133913,"source":"zoomeye","ip":"8.209.253.237","port":7707,"host":""}
[zoomeye] {"timestamp":1678133913,"source":"zoomeye","ip":"8.209.243.173","port":7707,"host":""}
[zoomeye] {"timestamp":1678133913,"source":"zoomeye","ip":"8.219.74.58","port":7707,"host":""}
[zoomeye] {"timestamp":1678133913,"source":"zoomeye","ip":"8.219.5.240","port":7707,"host":""}
[zoomeye] {"timestamp":1678133913,"source":"zoomeye","ip":"8.208.89.32","port":7707,"host":""}
[zoomeye] {"timestamp":1678133913,"source":"zoomeye","ip":"8.219.43.134","port":7707,"host":""}
[zoomeye] {"timestamp":1678133913,"source":"zoomeye","ip":"37.143.129.214","port":9292,"host":""}
[zoomeye] {"timestamp":1678133913,"source":"zoomeye","ip":"8.209.240.66","port":7707,"host":""}
[zoomeye] {"timestamp":1678133913,"source":"zoomeye","ip":"8.219.169.172","port":7707,"host":""}
[zoomeye] {"timestamp":1678133913,"source":"zoomeye","ip":"8.213.128.6","port":7707,"host":""}
[zoomeye] {"timestamp":1678133913,"source":"zoomeye","ip":"157.175.44.122","port":8002,"host":""}
[zoomeye] {"timestamp":1678133913,"source":"zoomeye","ip":"188.121.114.63","port":8880,"host":""}
[zoomeye] {"timestamp":1678133913,"source":"zoomeye","ip":"188.121.119.124","port":8880,"host":""}
[zoomeye] {"timestamp":1678133913,"source":"zoomeye","ip":"188.121.116.204","port":8880,"host":""}
[zoomeye] {"timestamp":1678133913,"source":"zoomeye","ip":"93.240.86.179","port":8182,"host":""}
[zoomeye] {"timestamp":1678133913,"source":"zoomeye","ip":"93.240.182.242","port":8182,"host":""}
[zoomeye] {"timestamp":1678133913,"source":"zoomeye","ip":"38.242.7.18","port":3128,"host":""}
```

### 查询日本所有数据库
```
./uncover-turbo -v -q '日本地区所有数据库' -json -delay 5 -r

  __  ______  _________ _   _____  _____
 / / / / __ \/ ___/ __ \ | / / _ \/ ___/
/ /_/ / / / / /__/ /_/ / |/ /  __/ /    
\__,_/_/ /_/\___/\____/|___/\___/_/ v1.0.2

                projectdiscovery.io

[DBG] Translate to fofa query: country="JP" && app="MySQL" or app="PostgreSQL" or app="MongoDB" or app="Redis" or app="Elasticsearch"
[DBG] Translate to quake query: 
[DBG] Translate to zoomeye query: country:"JP" && app:"mongodb"
[DBG] Translate to shodan query: country:"JP" port:27017,3306,5432 product:"MongoDB" or product:"MySQL" or product:"PostgreSQL"
```
