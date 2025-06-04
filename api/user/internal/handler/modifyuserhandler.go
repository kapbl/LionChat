package handler

import (
	"net/http"

	"chatLion/api/user/internal/logic"
	"chatLion/api/user/internal/svc"
	"chatLion/api/user/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ModifyUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ModifyUserRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewModifyUserLogic(r.Context(), svcCtx)
		resp, err := l.ModifyUser(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
