import { useState, useCallback } from 'react'
import toast from 'react-hot-toast'

export function useToast() {
  const [toasts, setToasts] = useState([])

  const showToast = useCallback((message, options = {}) => {
    const id = toast(message, options)
    return id
  }, [])

  const success = useCallback((message, options = {}) => {
    return toast.success(message, options)
  }, [])

  const error = useCallback((message, options = {}) => {
    return toast.error(message, options)
  }, [])

  const loading = useCallback((message, options = {}) => {
    return toast.loading(message, options)
  }, [])

  const dismiss = useCallback((id) => {
    toast.dismiss(id)
  }, [])

  const promise = useCallback((promise, messages) => {
    return toast.promise(promise, messages)
  }, [])

  return {
    success,
    error,
    loading,
    dismiss,
    promise,
    show: showToast,
  }
}