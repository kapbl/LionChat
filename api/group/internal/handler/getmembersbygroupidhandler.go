package handler

import (
	"net/http"

	"chatLion/api/group/internal/logic"
	"chatLion/api/group/internal/svc"
	"chatLion/api/group/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetMembersByGroupIDHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetMemberRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewGetMembersByGroupIDLogic(r.Context(), svcCtx)
		resp, err := l.GetMembersByGroupID(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
