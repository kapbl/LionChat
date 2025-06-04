package handler

import (
	"net/http"

	"chatLion/api/friend/internal/logic"
	"chatLion/api/friend/internal/svc"
	"chatLion/api/friend/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetFriendsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetFriendsRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewGetFriendsLogic(r.Context(), svcCtx)
		resp, err := l.GetFriends(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
