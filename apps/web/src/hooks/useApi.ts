import { useState, useEffect } from 'react'
import api from '../lib/api'
import { fetchCategorySummary } from '../lib/api'
import type { CategorySummary, PaginatedResponse } from '../types'

type QueryParamValue = string | number | boolean | null | undefined

type QueryParams = Record<string, QueryParamValue>

function isPaginatedEnvelope<T>(value: unknown): value is PaginatedResponse<T> {
  if (typeof value !== 'object' || value === null) {
    return false
  }

  const candidate = value as Partial<PaginatedResponse<T>>
  return Array.isArray(candidate.items) && typeof candidate.total === 'number'
}

export function useApi<T>(url: string, deps: unknown[] = [], params?: QueryParams) {
  const [data, setData] = useState<T | null>(null)
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const normalizedParams = Object.entries(params ?? {}).reduce<Record<string, string | number | boolean>>(
    (acc, [key, value]) => {
      if (value !== undefined && value !== null) {
        acc[key] = value
      }
      return acc
    },
    {},
  )

  const serializedParams = JSON.stringify(normalizedParams)

  const fetch = async () => {
    try {
      setLoading(true)
      setError(null)
      const res = await api.get<PaginatedResponse<unknown> | T>(url, {
        params: normalizedParams,
      })

      if (isPaginatedEnvelope(res.data)) {
        setData(res.data.items as T)
        setTotal(res.data.total)
        setPage(res.data.page || 1)
        setTotalPages(res.data.total_pages || 1)
      } else {
        setData(res.data)
        setTotal(Array.isArray(res.data) ? res.data.length : 0)
        setPage(1)
        setTotalPages(1)
      }
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
  }, [url, serializedParams, ...deps])

  return { data, total, page, totalPages, loading, error, refetch: fetch }
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
