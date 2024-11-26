import { useState, useEffect } from 'react';

export interface Stack {
  id: string;
  name: string;
  status: 'in_progress' | 'review' | 'complete';
  branches: Branch[];
}

export interface Branch {
  name: string;
  status: 'ready' | 'review' | 'draft';
  commits: number;
  author: string;
  timeago: string;
  description: string;
  conflicts?: boolean;
  pr?: {
    number: number;
    status: 'open' | 'draft' | 'merged';
  };
}

export function useStacks(repoId: string) {
  const [stacks, setStacks] = useState<Stack[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    fetchStacks();
  }, [repoId])

  const fetchStacks = async () => {
    try {
      const response = await fetch(`/api/respositories/${repoId}/stacks`);
      const data = await response.json();
      setStacks(data);
    } catch (err) {
      setError(err as Error);
    } finally {
      setLoading(false);
    }
  };

  const createStack = async (name: string) => {
    // Implementation
  }

  const addBranch = async (stackId: string, branchName: string) => { };

  const syncStack = async (stackId: string) => { };

  return {
    stacks,
    loading,
    error,
    createStack,
    addBranch,
    syncStack,
    refresh: fetchStacks,
  }
}
