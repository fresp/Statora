import { useMemo } from 'react'
import { useSearchParams } from 'react-router-dom'

interface UseAdminPaginationOptions {
  pageParam?: string
  limitParam?: string
  defaultPage?: number
  defaultLimit?: number
}

function parsePositiveInteger(value: string | null, fallback: number) {
  if (!value) return fallback
  const parsed = Number.parseInt(value, 10)
  if (!Number.isFinite(parsed) || parsed < 1) return fallback
  return parsed
}

export function useAdminPagination(options: UseAdminPaginationOptions = {}) {
  const {
    pageParam = 'page',
    limitParam = 'limit',
    defaultPage = 1,
    defaultLimit = 10,
  } = options

  const [searchParams, setSearchParams] = useSearchParams()

  const page = parsePositiveInteger(searchParams.get(pageParam), defaultPage)
  const limit = parsePositiveInteger(searchParams.get(limitParam), defaultLimit)

  const setPage = (nextPage: number) => {
    const resolvedPage = Math.max(1, nextPage)
    const next = new URLSearchParams(searchParams)
    next.set(pageParam, String(resolvedPage))
    next.set(limitParam, String(limit))
    setSearchParams(next)
  }

  const setLimit = (nextLimit: number) => {
    const resolvedLimit = Math.max(1, nextLimit)
    const next = new URLSearchParams(searchParams)
    next.set(limitParam, String(resolvedLimit))
    next.set(pageParam, String(defaultPage))
    setSearchParams(next)
  }

  const apiParams = useMemo(
    () => ({ page, limit }),
    [page, limit],
  )

  return {
    page,
    limit,
    apiParams,
    setPage,
    setLimit,
  }
}
