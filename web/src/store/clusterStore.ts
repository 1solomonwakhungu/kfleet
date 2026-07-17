import { create } from 'zustand';
import type { Cluster, ClusterUpdate } from '@/types/cluster';

interface ClusterStoreState {
  clusters: Record<string, Cluster>;
  setClusters: (clusters: Cluster[]) => void;
  upsertCluster: (cluster: Cluster) => void;
  removeCluster: (id: string) => void;
  applyUpdate: (update: ClusterUpdate) => void;
}

export const useClusterStore = create<ClusterStoreState>((set) => ({
  clusters: {},

  setClusters: (clusters) => set({ clusters: Object.fromEntries(clusters.map((c) => [c.id, c])) }),

  upsertCluster: (cluster) => set((state) => ({ clusters: { ...state.clusters, [cluster.id]: cluster } })),

  removeCluster: (id) =>
    set((state) => {
      const next = { ...state.clusters };
      delete next[id];
      return { clusters: next };
    }),

  applyUpdate: (update) =>
    set((state) => {
      if (update.type === 'deleted') {
        const next = { ...state.clusters };
        delete next[update.cluster.id];
        return { clusters: next };
      }
      return { clusters: { ...state.clusters, [update.cluster.id]: update.cluster } };
    }),
}));
