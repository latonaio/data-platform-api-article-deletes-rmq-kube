# data-platform-api-article-deletes-rmq-kube
data-platform-api-article-deletes-rmq-kube は、周辺システム　を データ連携基盤 と統合することを目的に、API で記事データに削除フラグを設定するマイクロサービスです。  
https://xxx.xxx.io/api/API_ARTICLE_SRV/deletes/

## 動作環境
data-platform-api-article-deletes-rmq-kube の動作環境は、次の通りです。  
・ OS: LinuxOS （必須）  
・ CPU: ARM/AMD/Intel（いずれか必須）  

## 本レポジトリ が 対応する API サービス
data-platform-api-article-deletes-rmq-kube が対応する APIサービス は、次のものです。

APIサービス URL: https://xxx.xxx.io/api/API_ARTICLE_SRV/deletes/

## 本レポジトリ に 含まれる API名
data-platform-api-article-deletes-rmq-kube には、次の API をコールするためのリソースが含まれています。  

* A_Header（記事 - ヘッダ）

## API への 値入力条件 の 初期値
data-platform-api-article-deletes-rmq-kube において、API への値入力条件の初期値は、入力ファイルレイアウトの種別毎に、次の通りとなっています。  

## データ連携基盤のAPIの選択的コール
Latona および AION の データ連携基盤 関連リソースでは、Inputs フォルダ下の sample.json の accepter に削除フラグ設定したいデータの種別（＝APIの種別）を入力し、指定することができます。  
なお、同 accepter にAll(もしくは空白)の値を入力することで、全データ（＝全APIの種別）をまとめて削除フラグ設定することができます。  

* sample.jsonの記載例(1)  

accepter において 下記の例のように、データの種別（＝APIの種別）を指定します。  
ここでは、"Header" が指定されています。    
  
```
	"api_schema": "DPFMArticleDeletes",
	"accepter": ["Header"],
```
  
* 全データを取得する際のsample.jsonの記載例(2)  

全データを取得する場合、sample.json は以下のように記載します。  

```
	"api_schema": "DPFMArticleDeletes",
	"accepter": ["All"],
```

## 指定されたデータ種別のコール
accepter における データ種別 の指定に基づいて DPFM_API_Caller 内の caller.go で API がコールされます。  
caller.go の func() 毎 の 以下の箇所が、指定された API をコールするソースコードです。  

```
func (c *DPFMAPICaller) AsyncDeletes(
	accepter []string,
	input *dpfm_api_input_reader.SDC,
	output *dpfm_api_output_formatter.SDC,
	log *logger.Logger,
) (interface{}, []error) {
	wg := sync.WaitGroup{}
	mtx := sync.Mutex{}
	errs := make([]error, 0, 5)
	exconfAllExist := false

	exconfFin := make(chan error)
	subFuncFin := make(chan error)

	subfuncSDC := &sub_func_complementer.SDC{}

	wg.Add(1)
	go c.exconfProcess(&mtx, &wg, exconfFin, input, output, &exconfAllExist, &errs, log)
	if input.APIType == "deletes" {
		go c.subfuncProcess(&mtx, &wg, subFuncFin, input, output, subfuncSDC, accepter, &errs, log)
	} else {
		go func() { subFuncFin <- nil }()
	}

	ticker := time.NewTicker(10 * time.Second)
	if err := c.finWait(&mtx, exconfFin, ticker); err != nil || len(errs) != 0 {
		if err != nil {
			errs = append(errs, err)
		}
		return subfuncSDC, errs
	}
	if !exconfAllExist {
		mtx.Lock()
		return subfuncSDC, nil
	}
	wg.Wait()
	for range accepter {
		if err := c.finWait(&mtx, subFuncFin, ticker); err != nil || len(errs) != 0 {
			if err != nil {
				errs = append(errs, err)
			}
			return subfuncSDC, errs
		}
	}
```

## Output  
本マイクロサービスでは、[golang-logging-library-for-data-platform](https://github.com/latonaio/golang-logging-library-for-data-platform) により、以下のようなデータがJSON形式で出力されます。  
以下の sample.json の例は 記事 の ヘッダデータ が削除フラグ設定された結果の JSON の例です。  
以下の項目のうち、"Article" ～ "IsMarkedForDeletion" は、/DPFM_API_Output_Formatter/type.go 内 の Type Header {} による出力結果です。"cursor" ～ "time"は、golang-logging-library による 定型フォーマットの出力結果です。  

```
XXXXXXXX
```
