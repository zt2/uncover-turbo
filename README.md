# 简介
---

这是一个有趣的小 demo，想测试是否可以通过 GPT-3.5 为众多语法繁杂的测绘搜索引擎建立一个 `巴别塔`，来实现通用的自然语言测绘引擎，打通自然语言到测绘语法的最后一公里。

项目基于 project-discovery 的 [uncover](https://github.com/projectdiscovery/uncover) 改造而成，加入了自然语言到查询语法的翻译。

注意：由 GPT-3.5 生成的测绘引擎语法常出现语法错误

## 怎么玩
---
1. 将 OpenAI Token 设置为环境变量 `OPENAI_KEY`
2. 按照官方指导正常配置 `uncover`
3. 直接传入自然语言作为查询语法，改造后的 `uncover` 会使用 GPT-3.5-turbo 尽可能的将输入翻译成 `shodan`, `fofa`, `quake`, `zoomeye` 等测绘引擎的语法：

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
[quake] {"timestamp":1678132746,"source":"quake","ip":"198.154.193.49","port":3306,"host":"49-193-154-198.unifiedlayer.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"107.22.236.55","port":3306,"host":"ec2-107-22-236-55.compute-1.amazonaws.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"52.0.217.223","port":3306,"host":"ec2-52-0-217-223.compute-1.amazonaws.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"34.192.89.105","port":3306,"host":"ec2-34-192-89-105.compute-1.amazonaws.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"3.33.249.243","port":3306,"host":"a1a68052cd1d5813d.awsglobalaccelerator.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"100.25.81.193","port":3306,"host":"ec2-100-25-81-193.compute-1.amazonaws.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"137.184.101.220","port":3306,"host":"coinsprime.net"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"34.192.89.188","port":3306,"host":"ec2-34-192-89-188.compute-1.amazonaws.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"69.16.252.12","port":3306,"host":"host.hostsearchindia.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"137.184.101.219","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"157.230.214.142","port":3306,"host":"ptr0.jobscareerdaily.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"3.217.243.50","port":3306,"host":"ec2-3-217-243-50.compute-1.amazonaws.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"154.87.59.182","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"69.16.252.44","port":3306,"host":"wells.snositesaso2.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"157.230.214.43","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"67.225.147.2","port":3306,"host":"host.serverresponse.net"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"198.154.193.252","port":3306,"host":"252-193-154-198.unifiedlayer.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"217.194.133.109","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"157.230.214.156","port":3306,"host":"server.majesticsoftware.net"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"67.225.147.117","port":3306,"host":"host.thesyncsystems.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"104.197.196.6","port":3306,"host":"6.196.197.104.bc.googleusercontent.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"137.184.101.118","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"104.197.196.185","port":3306,"host":"185.196.197.104.bc.googleusercontent.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"34.192.89.16","port":3306,"host":"ec2-34-192-89-16.compute-1.amazonaws.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"13.66.29.25","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"198.154.193.213","port":3306,"host":"198-154-193-213.unifiedlayer.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"137.184.101.128","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"137.184.101.154","port":3306,"host":"chr-6.49-ftpandvpnserver-1vcpu-1gb-nyc1-01"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"217.194.133.138","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"52.202.241.183","port":3306,"host":"ec2-52-202-241-183.compute-1.amazonaws.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"157.230.214.39","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"104.197.196.36","port":3306,"host":"36.196.197.104.bc.googleusercontent.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"67.225.147.47","port":3306,"host":"host.locotrips.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"69.16.252.43","port":3306,"host":"wells.snositesaso2.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"67.225.147.118","port":3306,"host":"host.secretsfl.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"104.197.196.63","port":3306,"host":"63.196.197.104.bc.googleusercontent.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"157.230.214.4","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"5.8.41.4","port":3306,"host":"kvm3.ny2.gcorelabs.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"69.16.252.20","port":3306,"host":"host.aekhost.co.in"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"67.225.147.89","port":3306,"host":"host.serverresponse.net"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"100.25.81.252","port":3306,"host":"ec2-100-25-81-252.compute-1.amazonaws.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"198.154.193.253","port":3306,"host":"198-154-193-253.unifiedlayer.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"65.110.76.250","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"107.170.77.187","port":3306,"host":"wikimapping.org"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"67.225.147.88","port":3306,"host":"host.serverresponse.net"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"67.225.147.87","port":3306,"host":"host.serverresponse.net"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"67.225.147.44","port":3306,"host":"host.locotrips.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"67.225.147.198","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"3.86.240.233","port":3306,"host":"ec2-3-86-240-233.compute-1.amazonaws.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"107.170.77.235","port":3306,"host":"torcu.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"34.215.85.117","port":3306,"host":"ec2-34-215-85-117.us-west-2.compute.amazonaws.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"3.231.103.142","port":3306,"host":"ec2-3-231-103-142.compute-1.amazonaws.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"67.231.254.196","port":3306,"host":"67-231-254-196.static.as40244.net"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"67.225.147.49","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"67.225.147.143","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"69.16.252.8","port":3306,"host":"host.aekhost.co.in"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"45.151.134.120","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"67.225.147.78","port":3306,"host":"host4.mkcourtreporting.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"67.225.147.153","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"67.225.147.179","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"179.61.15.166","port":3306,"host":"166.15.61.179.bog.host.as64114.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"52.6.72.144","port":3306,"host":"ec2-52-6-72-144.compute-1.amazonaws.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"67.225.147.182","port":3306,"host":"asphost214.asphostserver.org"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"67.225.147.48","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"67.225.147.51","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"67.225.147.38","port":3306,"host":"host.serverresponse.net"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"67.225.147.3","port":3306,"host":"host.worldwidemaritimemarketing.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"67.225.147.199","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"67.225.147.45","port":3306,"host":"host.locotrips.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"67.225.147.50","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"67.225.147.229","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"67.225.147.186","port":3306,"host":"asphost214.asphostserver.org"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"67.225.147.180","port":3306,"host":"asphost214.asphostserver.org"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"67.225.147.233","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"67.225.147.46","port":3306,"host":"host.locotrips.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"154.91.184.134","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"52.6.72.20","port":3306,"host":"ec2-52-6-72-20.compute-1.amazonaws.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"154.91.184.237","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"67.231.254.205","port":3306,"host":"67-231-254-205.static.as40244.net"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"52.91.152.59","port":3306,"host":"ec2-52-91-152-59.compute-1.amazonaws.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"54.174.91.87","port":3306,"host":"ec2-54-174-91-87.compute-1.amazonaws.com"}
[quake] {"timestamp":1678132746,"source":"quake","ip":"154.91.184.227","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"154.91.184.235","port":3306,"host":""}
[quake] {"timestamp":1678132746,"source":"quake","ip":"5.8.41.205","port":3306,"host":"kvm5.ny2.gcorelabs.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"67.231.254.205","port":3306,"host":"67-231-254-205.static.as40244.net"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"52.91.152.59","port":3306,"host":"ec2-52-91-152-59.compute-1.amazonaws.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"54.174.91.87","port":3306,"host":"ec2-54-174-91-87.compute-1.amazonaws.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"154.91.184.227","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"154.91.184.235","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"5.8.41.205","port":3306,"host":"kvm5.ny2.gcorelabs.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"154.91.184.139","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"65.110.76.251","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"154.91.184.232","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"154.91.184.234","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"5.8.41.233","port":3306,"host":"kvm12.ny2.gcorelabs.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"5.8.41.231","port":3306,"host":"kvm10.ny2.gcorelabs.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"154.91.184.226","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"5.8.41.240","port":3306,"host":"kvm6.ny2.gcorelabs.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"52.37.179.137","port":3306,"host":"ec2-52-37-179-137.us-west-2.compute.amazonaws.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"154.91.184.229","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"108.186.194.244","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"144.126.248.203","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"137.184.101.155","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"64.182.141.16","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"3.141.117.101","port":3306,"host":"ec2-3-141-117-101.us-east-2.compute.amazonaws.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"154.91.184.136","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"65.110.76.252","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"54.153.84.97","port":3306,"host":"ec2-54-153-84-97.us-west-1.compute.amazonaws.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"154.91.184.137","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"154.87.59.216","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"157.230.214.37","port":3306,"host":"mrisim.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"54.215.165.122","port":3306,"host":"ec2-54-215-165-122.us-west-1.compute.amazonaws.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"198.154.193.212","port":3306,"host":"212-193-154-198.unifiedlayer.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"157.230.214.197","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"54.215.165.208","port":3306,"host":"ec2-54-215-165-208.us-west-1.compute.amazonaws.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"137.184.101.172","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"179.61.15.164","port":3306,"host":"164.15.61.179.bog.host.as64114.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"44.228.22.17","port":3306,"host":"ec2-44-228-22-17.us-west-2.compute.amazonaws.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"157.230.214.168","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"65.110.76.253","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"154.91.184.130","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"44.228.22.1","port":3306,"host":"ec2-44-228-22-1.us-west-2.compute.amazonaws.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"179.61.15.165","port":3306,"host":"165.15.61.179.bog.host.as64114.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"154.91.184.228","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"179.61.15.158","port":3306,"host":"158.15.61.179.bog.host.as64114.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"179.61.15.152","port":3306,"host":"152.15.61.179.bog.host.as64114.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"64.182.141.10","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"44.194.164.101","port":3306,"host":"ec2-44-194-164-101.compute-1.amazonaws.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"132.145.175.232","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"64.182.141.8","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"135.148.32.19","port":3306,"host":"vps-e4147e20.vps.ovh.us"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"54.210.39.4","port":3306,"host":"ec2-54-210-39-4.compute-1.amazonaws.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"35.223.197.123","port":3306,"host":"123.197.223.35.bc.googleusercontent.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"135.148.32.33","port":3306,"host":"vps-b1da6a33.vps.ovh.us"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"34.168.198.168","port":3306,"host":"168.198.168.34.bc.googleusercontent.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.94.74.100","port":3306,"host":"23-94-74-100-host.colocrossing.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"52.202.153.13","port":3306,"host":"ec2-52-202-153-13.compute-1.amazonaws.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"3.131.10.104","port":3306,"host":"ec2-3-131-10-104.us-east-2.compute.amazonaws.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"3.129.134.55","port":3306,"host":"ec2-3-129-134-55.us-east-2.compute.amazonaws.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"205.178.136.34","port":3306,"host":"34.136.178.205.netsolvps.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"144.126.248.152","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"52.202.153.38","port":3306,"host":"ec2-52-202-153-38.compute-1.amazonaws.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"35.170.209.187","port":3306,"host":"ec2-35-170-209-187.compute-1.amazonaws.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.225.109.118","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"205.178.136.38","port":3306,"host":"38.136.178.205.netsolvps.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"205.178.136.37","port":3306,"host":"37.136.178.205.netsolvps.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"137.184.101.22","port":3306,"host":""}
[quake] {"timestamp":1678132750,"source":"quake","ip":"23.94.74.222","port":3306,"host":"23-94-74-222-host.colocrossing.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"35.192.220.193","port":3306,"host":"193.220.192.35.bc.googleusercontent.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"34.192.89.207","port":3306,"host":"ec2-34-192-89-207.compute-1.amazonaws.com"}
[quake] {"timestamp":1678132750,"source":"quake","ip":"205.178.136.7","port":3306,"host":"7.136.178.205.netsolvps.com"}
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
