const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
const API_PATH = `${API_BASE}/api`;

export interface ApiError extends Error {
    status?: number;
    code?: string;
}

// Enhanced fetch wrapper with better error handling
export const apiRequest = async <T>(
    endpoint: string, 
    options: RequestInit = {}
): Promise<T> => {
    const url = `${API_PATH}${endpoint}`;
    const controller = new AbortController();
    
    // Set timeout for requests
    const timeoutId = setTimeout(() => controller.abort(), 30000);
    
    try {
        const response = await fetch(url, {
            ...options,
            signal: controller.signal,
            headers: {
                'Content-Type': 'application/json',
                ...options.headers,
            },
        });

        clearTimeout(timeoutId);

        if (!response.ok) {
            let errorMessage = `HTTP ${response.status}: ${response.statusText}`;
            
            try {
                const errorData = await response.json();
                errorMessage = errorData.message || errorData.error || errorMessage;
            } catch {
                // If we can't parse JSON, use the default message
            }

            const error = new Error(errorMessage) as ApiError;
            error.status = response.status;
            throw error;
        }

        // Handle empty responses
        const contentType = response.headers.get('content-type');
        if (contentType && contentType.includes('application/json')) {
            return await response.json();
        }
        
        return await response.text() as unknown as T;
        
    } catch (error) {
        clearTimeout(timeoutId);
        
        if (error instanceof Error) {
            if (error.name === 'AbortError') {
                throw new Error('Request timed out. Please try again.');
            }
            
            // Network errors
            if (!navigator.onLine) {
                throw new Error('No internet connection. Please check your network.');
            }
            
            // Connection refused or server down
            if (error.message.includes('fetch')) {
                throw new Error('Unable to connect to database server. Please ensure the backend is running.');
            }
        }
        
        throw error;
    }
};

// Retry logic for failed requests
export const withRetry = async <T>(
    operation: () => Promise<T>,
    maxRetries: number = 3,
    delay: number = 1000
): Promise<T> => {
    let lastError: Error;
    
    for (let attempt = 1; attempt <= maxRetries; attempt++) {
        try {
            return await operation();
        } catch (error) {
            lastError = error instanceof Error ? error : new Error('Unknown error');
            
            // Don't retry on client errors (4xx)
            if ('status' in lastError && typeof lastError.status === 'number' && lastError.status >= 400 && lastError.status < 500) {
                throw lastError;
            }
            
            if (attempt === maxRetries) {
                throw lastError;
            }
            
            // Wait before retrying
            await new Promise(resolve => setTimeout(resolve, delay * attempt));
        }
    }
    
    throw lastError!;
};
