package gitops

//
//func (gitops *Gitops) Refresh(request contracts.Control) contracts.Response {
//	gitopsObj := gitops.Shared.Registry.Find(request.Group, request.Name)
//
//	if gitopsObj != nil {
//
//		event := events.New(events.EVENT_REFRESH, static.KIND_GITOPS, gitopsObj.GetGroup(), gitopsObj.GetName(), nil)
//
//		bytes, err := event.ToJSON()
//
//		if err != nil {
//			return common.Response(http.StatusInternalServerError, static.RESPONSE_INTERNAL_ERROR, err, nil)
//		}
//
//		gitops.Shared.Manager.Cluster.KVStore.Propose(event.GetKey(), bytes, gitopsObj.Definition.GetRuntime().GetNode())
//		return common.Response(http.StatusOK, static.RESPONSE_REFRESHED, nil, nil)
//	} else {
//		return common.Response(http.StatusNotFound, static.RESPONSE_NOT_FOUND, nil, nil)
//	}
//}
//
//func (gitops *Gitops) Sync(request contracts.Control) contracts.Response {
//	gitopsObj := gitops.Shared.Registry.Find(request.Group, request.Name)
//
//	if gitopsObj != nil {
//		event := events.New(events.EVENT_SYNC, static.KIND_GITOPS, gitopsObj.GetGroup(), gitopsObj.GetName(), nil)
//
//		bytes, err := event.ToJSON()
//
//		if err != nil {
//			return common.Response(http.StatusInternalServerError, static.RESPONSE_INTERNAL_ERROR, err, nil)
//		}
//
//		gitops.Shared.Manager.Replication.EventsC <- KV.NewEncode(event.GetKey(), bytes, gitops.Shared.Manager.Config.KVStore.Node)
//		gitops.Shared.Manager.Cluster.KVStore.Propose(event.GetKey(), bytes, gitopsObj.Definition.GetRuntime().GetNode())
//
//		return common.Response(http.StatusOK, static.RESPONSE_SYNCED, nil, nil)
//	} else {
//		return common.Response(http.StatusNotFound, static.RESPONSE_NOT_FOUND, nil, nil)
//	}
//}
