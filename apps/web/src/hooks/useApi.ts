import { useState, useEffect } from 'react'
import api from '../lib/api'
import { fetchCategorySummary } from '../lib/api'
import type { CategorySummary } from '../types'

export function useApi<T>(url: string, deps: unknown[] = []) {
  const [data, setData] = useState<T | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetch = async () => {
    try {
      setLoading(true)
      setError(null)
      const res = await api.get<T>(url)
      setData(res.data)
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : 'Request failed'
      setError(msg)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetch()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, deps)

  return { data, loading, error, refetch: fetch }
}

export function useCategorySummary(categoryPrefix: string | undefined) {
  const [data, setData] = useState<CategorySummary | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetch = async () => {
    if (!categoryPrefix) {
      setData(null)
      setError('Missing category prefix')
      setLoading(false)
      return
    }

    try {
      setLoading(true)
      setError(null)
      const summary = await fetchCategorySummary(categoryPrefix)
      setData(summary)
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : 'Request failed'
      setError(msg)
      setData(null)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetch()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [categoryPrefix])

  return { data, loading, error, refetch: fetch }
}
