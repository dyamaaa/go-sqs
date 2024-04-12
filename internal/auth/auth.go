package auth

import (
	"net/http"
	"os"
)

// GetAuthToken は環境変数から認証トークンを取得します。
func GetAuthToken() string {
	return os.Getenv("AUTH_TOKEN")
}

// リクエストの "Authorization" ヘッダーに含まれるトークンが期待されるトークンと一致するかを確認します。
// 一致しない場合、レスポンスのステータスコードを "Unauthorized" に設定します。
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization") // リクエストからトークンを取得します。
		expectedToken := GetAuthToken()        // 期待されるトークンを取得します。

		if token != expectedToken { // トークンが一致しない場合、エラーレスポンスを返します。
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r) // トークンが一致する場合、次のハンドラーを実行します。
	})
}
