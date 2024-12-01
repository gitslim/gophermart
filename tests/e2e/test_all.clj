(ns test-all
  (:require [babashka.http-client :as http]
            [clojure.java.io :as io]
            [cheshire.core :as json]))

(def shop "http://localhost:8080/api/")
(def accrual "http://localhost:8081/api/")
(def !cookie (atom nil))
(def login (random-uuid))
(def password (random-uuid))
(def valid-luhn-order "49927398716")
(def invalid-luhn-order "79927398710")

(defn call [meth path data & {:keys [base]}]
  (let [api (or base shop)
        f (case meth
            :post http/post
            :put http/put
            http/get)
        cookie (or @!cookie "")
        [ct body] (if (string? data) ["text/plain" data] ["application/json" (json/encode data)])
        resp (try (f (str api path) {:headers {:content-type ct
                                               :cookie cookie}
                                     :body body})
                  (catch Exception e
                    (ex-data e)))
        status (:status resp)
        headers (:headers resp)
        data  (when (= status 200) (json/parse-string (:body resp)))]
    {:status status
     :headers headers
     :data data}))

(call :get "user/orders" nil)
(call :get "user/balance" nil)
(call :get "user/withdrawals" nil)
(call :post "user/balance/withdraw" {:order "2377225624" :sum 751})
(call :post "user/register" {:Login login :Password password})
(reset! !cookie (-> (call :post "user/login" {:Login login :Password password}) :headers (get "set-cookie")))
(call :post "user/orders" valid-luhn-order)
(call :post "user/orders" invalid-luhn-order)

(call :get (str "orders/" valid-luhn-order) nil :base accrual)
