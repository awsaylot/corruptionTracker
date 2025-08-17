// frontend/pages/extraction/index.tsx
import { NextPage } from 'next';
import { useState, useCallback } from 'react';
import MainLayout from '../../components/layout/MainLayout';
import URLInput from '../../components/extraction/URLInput';
import AnalysisProgress from '../../components/extraction/AnalysisProgress';
import ExtractedResults from '../../components/extraction/ExtractedResults';

interface ExtractionSession {
  sessionId: string;
  article: {
    id: string;
    url: string;
    title: string;
    source: string;
    author: string;
    publishDate: string;
  };
  status: string;
}

interface AnalysisSession {
  id: string;
  articleId: string;
  stages: any[];
  currentStage: number;
  status: 'running' | 'completed' | 'failed' | 'terminated';
  startedAt: string;
  completedAt?: string;
  finalResults?: any;
  evidence?: any[];
  hypotheses?: any[];
}

const ExtractionPage: NextPage = () => {
  const [currentSession, setCurrentSession] = useState<ExtractionSession | null>(null);
  const [completedSession, setCompletedSession] = useState<AnalysisSession | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleURLSubmit = useCallback(async (url: string, depth: number) => {
    setLoading(true);
    setError(null);
    setCurrentSession(null);
    setCompletedSession(null);

    try {
      const response = await fetch('/api/extraction/url', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ url, depth }),
      });

      if (!response.ok) {
        const errorData = await response.text();
        throw new Error(errorData || 'Failed to start extraction');
      }

      const sessionData = await response.json();
      setCurrentSession(sessionData);
    } catch (err) {
      console.error('Extraction error:', err);
      setError(err instanceof Error ? err.message : 'Unknown error occurred');
    } finally {
      setLoading(false);
    }
  }, []);

  const handleAnalysisComplete = useCallback(async (session: AnalysisSession) => {
    setCompletedSession(session);
    
    // Fetch additional data (hypotheses and evidence)
    try {
      const [hypothesesRes, evidenceRes] = await Promise.all([
        fetch(`/api/extraction/hypotheses?sessionId=${session.id}`),
        fetch(`/api/extraction/evidence?sessionId=${session.id}`)
      ]);

      if (hypothesesRes.ok && evidenceRes.ok) {
        const hypothesesData = await hypothesesRes.json();
        const evidenceData = await evidenceRes.json();
        
        setCompletedSession(prev => prev ? {
          ...prev,
          hypotheses: hypothesesData.hypotheses || [],
          evidence: evidenceData.evidence || []
        } : null);
      }
    } catch (err) {
      console.error('Error fetching additional analysis data:', err);
    }
  }, []);

  const handleAnalysisError = useCallback((errorMessage: string) => {
    setError(errorMessage);
    setCurrentSession(null);
  }, []);

  const handleStartNew = useCallback(() => {
    setCurrentSession(null);
    setCompletedSession(null);
    setError(null);
  }, []);

  return (
    <MainLayout>
      <div className="container mx-auto px-4 py-8 space-y-8">
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold text-gray-900 mb-4">
            Corruption Analysis Engine
          </h1>
          <p className="text-lg text-gray-600 max-w-3xl mx-auto">
            Advanced sequential analysis of news articles to identify corruption patterns, 
            extract entities, map relationships, and generate investigative hypotheses.
          </p>
        </div>

        {/* Error Display */}
        {error && (
          <div className="max-w-4xl mx-auto">
            <div className="bg-red-50 border border-red-200 rounded-lg p-4">
              <div className="flex items-start">
                <div className="flex-shrink-0">
                  <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                  </svg>
                </div>
                <div className="ml-3">
                  <h3 className="text-sm font-medium text-red-800">Analysis Error</h3>
                  <div className="mt-2 text-sm text-red-700">
                    <p>{error}</p>
                  </div>
                  <div className="mt-4">
                    <button
                      onClick={handleStartNew}
                      className="bg-red-100 text-red-800 px-3 py-1 rounded text-sm hover:bg-red-200 transition-colors"
                    >
                      Start New Analysis
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* URL Input - Show when no active session */}
        {!currentSession && !completedSession && !loading && (
          <URLInput onSubmit={handleURLSubmit} isLoading={loading} />
        )}

        {/* Analysis Progress - Show during active analysis */}
        {currentSession && !completedSession && (
          <div className="space-y-6">
            {/* Article Info */}
            <div className="max-w-4xl mx-auto bg-blue-50 border border-blue-200 rounded-lg p-4">
              <h3 className="font-medium text-blue-900 mb-2">Analyzing Article:</h3>
              <p className="text-blue-800 font-medium">{currentSession.article.title}</p>
              <div className="mt-2 text-sm text-blue-700 space-x-4">
                <span>Source: {currentSession.article.source}</span>
                {currentSession.article.author && (
                  <span>Author: {currentSession.article.author}</span>
                )}
                <span>Date: {new Date(currentSession.article.publishDate).toLocaleDateString()}</span>
              </div>
            </div>
            
            <AnalysisProgress
              sessionId={currentSession.sessionId}
              onComplete={handleAnalysisComplete}
              onError={handleAnalysisError}
            />
          </div>
        )}

        {/* Completed Analysis Results */}
        {completedSession && completedSession.finalResults && (
          <div className="space-y-6">
            {/* Analysis Summary */}
            <div className="max-w-4xl mx-auto bg-green-50 border border-green-200 rounded-lg p-4">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="font-medium text-green-900 mb-2">Analysis Complete</h3>
                  <p className="text-green-800">
                    Processed {completedSession.stages.length} stages with overall confidence of{' '}
                    {completedSession.finalResults.confidence 
                      ? (completedSession.finalResults.confidence * 100).toFixed(1) + '%'
                      : 'N/A'
                    }
                  </p>
                  <div className="mt-2 text-sm text-green-700">
                    Duration: {
                      completedSession.completedAt && completedSession.startedAt
                        ? Math.round((new Date(completedSession.completedAt).getTime() - new Date(completedSession.startedAt).getTime()) / 1000) + ' seconds'
                        : 'N/A'
                    }
                  </div>
                </div>
                <button
                  onClick={handleStartNew}
                  className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
                >
                  New Analysis
                </button>
              </div>
            </div>

            <ExtractedResults
              results={completedSession.finalResults}
              hypotheses={completedSession.hypotheses}
              evidence={completedSession.evidence}
            />
          </div>
        )}

        {/* Loading State */}
        {loading && !currentSession && (
          <div className="max-w-4xl mx-auto text-center py-12">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
            <p className="text-gray-600">Initializing analysis...</p>
          </div>
        )}

        {/* Feature Info */}
        {!currentSession && !completedSession && !loading && !error && (
          <div className="max-w-6xl mx-auto mt-16">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
              <div className="text-center p-6">
                <div className="w-12 h-12 bg-blue-100 rounded-lg flex items-center justify-center mx-auto mb-4">
                  <svg className="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                </div>
                <h3 className="text-lg font-semibold text-gray-900 mb-2">Sequential Analysis</h3>
                <p className="text-gray-600">
                  Multi-stage analysis pipeline that progressively extracts deeper insights from corruption-related news articles.
                </p>
              </div>

              <div className="text-center p-6">
                <div className="w-12 h-12 bg-green-100 rounded-lg flex items-center justify-center mx-auto mb-4">
                  <svg className="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                  </svg>
                </div>
                <h3 className="text-lg font-semibold text-gray-900 mb-2">Smart Entity Extraction</h3>
                <p className="text-gray-600">
                  Identifies people, organizations, locations, monetary amounts, and temporal information with confidence scoring.
                </p>
              </div>

              <div className="text-center p-6">
                <div className="w-12 h-12 bg-purple-100 rounded-lg flex items-center justify-center mx-auto mb-4">
                  <svg className="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                  </svg>
                </div>
                <h3 className="text-lg font-semibold text-gray-900 mb-2">Hypothesis Generation</h3>
                <p className="text-gray-600">
                  Generates investigative hypotheses, identifies information gaps, and suggests follow-up research directions.
                </p>
              </div>
            </div>

            <div className="mt-12 text-center">
              <h3 className="text-xl font-semibold text-gray-900 mb-4">Analysis Depth Levels</h3>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                  <h4 className="font-medium text-blue-900 mb-2">Basic (2-3)</h4>
                  <p className="text-sm text-blue-800">
                    Surface extraction and basic relationship mapping. Fast processing for quick insights.
                  </p>
                </div>
                <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
                  <h4 className="font-medium text-yellow-900 mb-2">Standard (4-6)</h4>
                  <p className="text-sm text-yellow-800">
                    Deep analysis with pattern recognition, cross-referencing, and corruption indicator detection.
                  </p>
                </div>
                <div className="bg-red-50 border border-red-200 rounded-lg p-4">
                  <h4 className="font-medium text-red-900 mb-2">Deep (7-10)</h4>
                  <p className="text-sm text-red-800">
                    Comprehensive analysis with hypothesis generation, evidence chains, and recursive refinement.
                  </p>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </MainLayout>
  );
};

export default ExtractionPage;