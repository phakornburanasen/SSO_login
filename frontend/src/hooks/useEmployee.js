import { useEffect, useState } from 'react'
import { api } from '../api.js'

/**
 * useEmployee — ดึงข้อมูลชื่อ-นามสกุล จาก /api/employee/{username}
 * คืน { employee, loading, error }
 *
 *   employee = {
 *     empId, form_first_name, form_last_name, fullName,
 *     title, department, position, email, profileImage,
 *     available, fromCache
 *   }
 */
export function useEmployee(username) {
  const [state, setState] = useState({ employee: null, loading: true, error: null })

  useEffect(() => {
    if (!username) {
      setState({ employee: null, loading: false, error: null })
      return
    }
    let cancelled = false
    setState((s) => ({ ...s, loading: true, error: null }))

    api.getEmployee(username)
      .then((data) => {
        if (cancelled) return
        setState({ employee: data, loading: false, error: null })
      })
      .catch((err) => {
        if (cancelled) return
        setState({ employee: null, loading: false, error: err.message || 'fetch failed' })
      })

    return () => { cancelled = true }
  }, [username])

  return state
}
