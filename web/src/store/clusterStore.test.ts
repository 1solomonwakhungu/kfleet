import { beforeEach, describe, expect, it } from 'vitest';
import type { Cluster } from '@/types/cluster';
import { useClusterStore } from './clusterStore';

const cluster = (id: string, name = id): Cluster => ({
  id,
  name,
  health: 'healthy',
  nodeCount: 1,
  podCount: 2,
  k8sVersion: '1.31',
  agentVersion: '0.1',
  lastHeartbeat: '2026-07-19T12:00:00Z',
  registeredAt: '2026-07-19T11:00:00Z',
  labels: {},
});

describe('clusterStore', () => {
  beforeEach(() => {
    useClusterStore.setState({ clusters: {} });
  });

  it('upserts clusters while preserving other records', () => {
    const first = cluster('one');
    const second = cluster('two');
    useClusterStore.getState().setClusters([first, second]);

    const updated = { ...first, name: 'One updated', podCount: 9 };
    useClusterStore.getState().upsertCluster(updated);

    expect(useClusterStore.getState().clusters).toEqual({ one: updated, two: second });
  });

  it('applies added and updated events as upserts', () => {
    const original = cluster('one');
    const updated = { ...original, health: 'degraded' as const };

    useClusterStore.getState().applyUpdate({ type: 'added', cluster: original });
    useClusterStore.getState().applyUpdate({ type: 'updated', cluster: updated });

    expect(useClusterStore.getState().clusters).toEqual({ one: updated });
  });

  it('deletes clusters through direct and WebSocket updates', () => {
    const first = cluster('one');
    const second = cluster('two');
    useClusterStore.getState().setClusters([first, second]);

    useClusterStore.getState().removeCluster('one');
    useClusterStore.getState().applyUpdate({ type: 'deleted', cluster: second });

    expect(useClusterStore.getState().clusters).toEqual({});
  });
});
