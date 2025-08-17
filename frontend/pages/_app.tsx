// pages/_app.tsx - Updated with error boundary and global state
import '../styles/globals.css'
import type { AppProps } from 'next/app'
import { useRouter } from 'next/router'
import { useEffect, useState } from 'react'
import MainLayout from '../components/layout/MainLayout'
import ErrorBoundary from '../components/common/ErrorBoundary'
import LoadingSpinner from '../components/ui/LoadingSpinner'
import { api } from '../utils/api'

export default function App({ Component, pageProps }: AppProps) {
  const router = useRouter()
  const [isRouting, setIsRouting] = useState(false)
  const [apiStatus, setApiStatus] = useState<'checking' | 'connected' | 'error'>('checking')

  // Check API connection on app start
  useEffect(() => {
    checkApiConnection()
  }, [])

  const checkApiConnection = async () => {
    try {
      await api.getAllNodes()
      setApiStatus('connected')
    } catch (error) {
      console.error('API connection failed:', error)
      setApiStatus('error')
    }
  }

  // Handle route change loading states
  useEffect(() => {
    const handleRouteChangeStart = () => setIsRouting(true)
    const handleRouteChangeComplete = () => setIsRouting(false)
    const handleRouteChangeError = () => setIsRouting(false)

    router.events.on('routeChangeStart', handleRouteChangeStart)
    router.events.on('routeChangeComplete', handleRouteChangeComplete)
    router.events.on('routeChangeError', handleRouteChangeError)

    return () => {
      router.events.off('routeChangeStart', handleRouteChangeStart)
      router.events.off('routeChangeComplete', handleRouteChangeComplete)
      router.events.off('routeChangeError', handleRouteChangeError)
    }
  }, [router.events])

  // Show API connection error
  if (apiStatus === 'error') {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center max-w-md mx-auto">
          <div className="bg-red-50 border-l-4 border-red-400 p-4 mb-6">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <h3 className="text-sm font-medium text-red-800">Database Connection Failed</h3>
                <div className="mt-2 text-sm text-red-700">
                  <p>Unable to connect to the graph database. Please ensure the backend server is running on localhost:8080.</p>
                </div>
              </div>
            </div>
          </div>
          
          <div className="space-y-3">
            <button
              onClick={checkApiConnection}
              className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            >
              Retry Connection
            </button>
            
            <div className="text-sm text-gray-600">
              <p>Make sure your backend server is running with:</p>
              <code className="block mt-2 p-2 bg-gray-100 rounded text-xs">
                npm run dev
              </code>
            </div>
          </div>
        </div>
      </div>
    )
  }

  // Show initial loading while checking API
  if (apiStatus === 'checking') {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <LoadingSpinner size="large" text="Connecting to database..." />
      </div>
    )
  }

  return (
    <ErrorBoundary>
      <MainLayout>
        {/* Route Loading Overlay */}
        {isRouting && (
          <div className="fixed inset-0 bg-white bg-opacity-75 flex items-center justify-center z-50">
            <LoadingSpinner size="large" text="Loading page..." />
          </div>
        )}
        
        <Component {...pageProps} />
      </MainLayout>
    </ErrorBoundary>
  )
}