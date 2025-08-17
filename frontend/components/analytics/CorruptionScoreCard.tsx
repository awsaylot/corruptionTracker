import React from 'react';
import { CorruptionScore } from '../../utils/api';

interface CorruptionScoreCardProps {
    score: CorruptionScore;
    isLoading?: boolean;
    error?: string;
}

const CorruptionScoreCard: React.FC<CorruptionScoreCardProps> = ({ score, isLoading, error }) => {
    if (isLoading) {
        return (
            <div className="bg-white shadow rounded-lg p-6 animate-pulse">
                <div className="h-4 bg-gray-200 rounded w-3/4 mb-4"></div>
                <div className="h-24 bg-gray-100 rounded mb-4"></div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="bg-red-50 border border-red-200 rounded-lg p-6">
                <h3 className="text-red-800 font-medium">Error Loading Score</h3>
                <p className="text-red-600 text-sm mt-2">{error}</p>
            </div>
        );
    }

    const getRiskColor = (level: string) => {
        switch (level.toLowerCase()) {
            case 'high':
                return 'bg-red-100 text-red-800';
            case 'medium':
                return 'bg-yellow-100 text-yellow-800';
            case 'low':
                return 'bg-green-100 text-green-800';
            default:
                return 'bg-gray-100 text-gray-800';
        }
    };

    return (
        <div className="bg-white shadow rounded-lg p-6">
            <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-medium text-gray-900">Corruption Risk Assessment</h3>
                <span className={`px-3 py-1 rounded-full text-sm font-medium ${getRiskColor(score.risk_level)}`}>
                    {score.risk_level} Risk
                </span>
            </div>

            <div className="space-y-4">
                <div>
                    <div className="flex items-center justify-between">
                        <span className="text-sm font-medium text-gray-500">Overall Score</span>
                        <span className="text-lg font-semibold text-gray-900">
                            {(score.corruption_score * 100).toFixed(1)}%
                        </span>
                    </div>
                    <div className="mt-2 w-full bg-gray-200 rounded-full h-2">
                        <div
                            className="bg-blue-600 rounded-full h-2"
                            style={{ width: `${score.corruption_score * 100}%` }}
                        />
                    </div>
                </div>

                <div className="grid grid-cols-2 gap-4">
                    {Object.entries(score.score_breakdown).map(([key, value]) => (
                        <div key={key} className="bg-gray-50 rounded p-3">
                            <div className="text-sm font-medium text-gray-500 capitalize">
                                {key.replace('_', ' ')}
                            </div>
                            <div className="mt-1 text-lg font-semibold text-gray-900">
                                {(value * 100).toFixed(1)}%
                            </div>
                        </div>
                    ))}
                </div>

                <div className="border-t pt-4 mt-4">
                    <div className="flex items-center justify-between text-sm">
                        <span className="text-gray-500">Last Updated</span>
                        <span className="text-gray-900">{new Date(score.updated_at).toLocaleDateString()}</span>
                    </div>
                    <div className="flex items-center justify-between text-sm mt-2">
                        <span className="text-gray-500">Total Mentions</span>
                        <span className="text-gray-900">{score.mention_count}</span>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default CorruptionScoreCard;
