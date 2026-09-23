# 第2章 JDBC（クライアントからDBを使うAPI）

## 基本のJDBC

- 5つのインタフェースで構成される：`Driver`, `Connection`, `Statement`, `ResultSet`, `ResultSetMetaData`
- ベンダー固有の部分はドライバクラスと接続文字列だけで、残りはベンダー非依存のコードで書ける。
- Connection と ResultSet はリソースを保持するので、使い終わったらすぐ close する。
- どのメソッドも `SQLException` を投げうるので、例外処理が必須。
- `ResultSetMetaData` で出力テーブルのスキーマ（列名・型・表示幅）が取れる。SQLインタプリタのように任意のクエリを受け付けるときに便利。

## 発展的なJDBC

- **DriverManager / DataSource**：接続の詳細を隠して、よりベンダー非依存にする。DataSource はドライバと接続文字列をまとめて持つ。
- **トランザクション**
  - デフォルトは autocommit で、1文が1トランザクションになる。
  - 重要な処理では `setAutoCommit(false)` にし、`commit` / `rollback` を明示的に呼ぶ。SQLが失敗したらロールバックする。
- **分離レベル（isolation level）**

  | レベル | 起こりうる問題 |
  |---|---|
  | Read-Uncommitted | 未コミット読み、非再現読み、ファントム |
  | Read-Committed | 非再現読み、ファントム |
  | Repeatable-Read | ファントムのみ |
  | Serializable | なし（ただし遅い） |

  Serializable が理想だが遅いので、リスクを分析したうえで緩いレベルを選ぶ。
- **PreparedStatement**：プレースホルダ付きのSQLを事前にコンパイルしておける。ループで繰り返し実行するときに効率がよい。
- **スクロール可能・更新可能な ResultSet**：デフォルトは前方向のみで更新不可。`createStatement` の引数で指定する。

## Java で計算するか、SQL で計算するか

- 経験則は「**エンジンにできるだけ仕事をさせる**」。欲しいデータをぴったり返すSQLを1本書いて、エンジンに任せる。
- クライアント側で自前のjoinなどをするのは早すぎる最適化になる。

## Go での対応

JDBC は Go の `database/sql` / `database/sql/driver` に相当する。

| JDBC | Go |
|---|---|
| `Driver` | `driver.Driver` |
| `Connection` | `sql.DB` / `sql.Conn` |
| `Statement` / `PreparedStatement` | `sql.Stmt` |
| `ResultSet` | `sql.Rows` |
| `ResultSetMetaData` | `Rows.ColumnTypes()` |
| `setAutoCommit(false)` / 分離レベル | `sql.Tx` / `sql.TxOptions.Isolation` |

SimpleDB の JDBC 実装部分をGoで書くなら、`database/sql/driver` のインタフェースを実装する形にすると自然。
