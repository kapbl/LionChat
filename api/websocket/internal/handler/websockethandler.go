package handler

import (
	"net/http"

	"chatLion/api/websocket/internal/logic"
	"chatLion/api/websocket/internal/svc"
	"chatLion/api/websocket/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func websocketHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.WebSocketRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewWebsocketLogic(r.Context(), svcCtx, w, r)
		resp, err := l.Websocket(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
