package main

import (
  "smr/pkg/definitions"
  "smr/pkg/manager"
  "smr/pkg/replicas"
)

func (implementation *Implementation) Implementation(mgr *manager.Manager, group string, containerDefinition definitions.Definition) ([]string, []string) {
  _, index := mgr.Registry.Name(group, mgr.Runtime.PROJECT)

  r := replicas.Replicas{
    Group:          group,
    GeneratedIndex: index,
  }

  groups, names := r.HandleReplica(mgr, containerDefinition)

  return groups, names
}

// Exported
var Container Implementation
