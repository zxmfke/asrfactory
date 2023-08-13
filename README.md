### ASR FACTORY

---

这个包主要是封装，阿里、快商通、百度、字节、腾讯和讯飞6家的`短语音识别`和`实时流语音识别`。

之前刚好在测试各家的语音识别相关功能，但是每家的返回值都不同，调用方式都不同，所以就封装了这么一个包。主要就是用简易工厂模式封装了一下，可以用来内部做测试。

功能方面，只是单纯的返回识别结果，实时流也是，正常是要再返回时间戳的，不过各家在时间戳上更是五花八门，就之后有空再封装。

#### Road Map

- 添加识别结果的字级和句级时间戳
- 提供一个 web server 的调用方式
- 完善文档
- 配置接口调用的账号

---

有什么需求也欢迎讨论，另外，接口的app，账号需要自己去生成。

| 短语音识别 | URL                                                          |
| ---------- | ------------------------------------------------------------ |
| 阿里       | [智能语音交互RESTfulAPI（ROA）示例_智能语音交互-阿里云帮助中心 (aliyun.com)](https://help.aliyun.com/document_detail/92131.html?spm=a2c4g.432038.0.0.533f74cbWU1MuL#section-og9-qpl-2jq) |
| 快商通     | [快商通AI开放平台-短语音识别](https://aihc.shengwenyun.com/asr-short-md) |
| 百度       | [短语音识别标准版API - 语音技术 (baidu.com)](https://cloud.baidu.com/doc/SPEECH/s/Jlbxdezuf) |
| 腾讯       | [语音识别 一句话识别-一句话识别相关接口-API 中心-腾讯云 (tencent.cn)](https://cloud.tencent.cn/document/api/1093/35646) |
| 科大       | [语音听写_语音识别-讯飞开放平台 (xfyun.cn)](https://www.xfyun.cn/services/voicedictation) |
| 字节       | [一句话识别--语音技术-火山引擎 (volcengine.com)](https://www.volcengine.com/docs/6561/80816) |

| 实时流语音识别 | URL                                                          |
| -------------- | ------------------------------------------------------------ |
| 阿里           | [如何自行开发代码访问阿里语音服务_智能语音交互-阿里云帮助中心 (aliyun.com)](https://help.aliyun.com/document_detail/324262.htm?spm=a2c4g.432038.0.0.327574cbrQ3qQx#topic-2121083) |
| 快商通         | [快商通AI开放平台-实时语音识别](https://aihc.shengwenyun.com/asr-stream-md) |
| 百度           | [语音技术 (baidu.com)](https://ai.baidu.com/ai-doc/SPEECH/2k5dllqxj) |
| 腾讯           | [语音识别 实时语音识别（websocket）-API 文档-文档中心-腾讯云 (tencent.com)](https://cloud.tencent.com/document/product/1093/48982) |
| 科大           | [实时语音转写_实时语音识别服务-讯飞开放平台 (xfyun.cn)](https://www.xfyun.cn/services/rtasr) |
| 字节           | [流式语音识别--语音技术-火山引擎 (volcengine.com)](https://www.volcengine.com/docs/6561/80818) |

