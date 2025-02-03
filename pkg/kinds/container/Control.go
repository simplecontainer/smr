package container

//func (container *Container) Restart(request contracts.Control) contracts.Response {
//	containerObj := container.Shared.Registry.Find(request.Group, request.Name)
//
//	if containerObj == nil {
//		return common.Response(http.StatusNotFound, static.STATUS_RESPONSE_NOT_FOUND, errors.New("container not found"), nil)
//	}
//
//	event := events.New(events.EVENT_RESTART, static.KIND_CONTAINER, static.KIND_CONTAINER, containerObj.GetGroup(), containerObj.GetGeneratedName(), nil)
//
//	bytes, err := event.ToJson()
//
//	if err != nil {
//		return common.Response(http.StatusInternalServerError, static.STATUS_RESPONSE_INTERNAL_ERROR, err, nil)
//	}
//
//	container.Shared.Manager.Replication.EventsC <- KV.NewEncode(event.GetKey(), bytes, container.Shared.Manager.Config.KVStore.Node)
//	return common.Response(http.StatusOK, static.STATUS_RESPONSE_RESTART, nil, nil)
//}
