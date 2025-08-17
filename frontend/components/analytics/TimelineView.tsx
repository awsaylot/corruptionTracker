import React from 'react';

interface TimelineEvent {
    id: string;
    date: string;
    title: string;
    description: string;
    type: string;
    entities: string[];
    importance: number;
}

interface TimelineViewProps {
    events: TimelineEvent[];
    isLoading?: boolean;
    error?: string;
    onEventClick?: (event: TimelineEvent) => void;
}

const TimelineView: React.FC<TimelineViewProps> = ({ 
    events, 
    isLoading, 
    error,
    onEventClick 
}) => {
    if (isLoading) {
        return (
            <div className="space-y-4 p-4">
                {[1, 2, 3].map((i) => (
                    <div key={i} className="animate-pulse flex space-x-4">
                        <div className="w-24 bg-gray-200 h-6 rounded"></div>
                        <div className="flex-1 space-y-2">
                            <div className="h-4 bg-gray-200 rounded w-3/4"></div>
                            <div className="h-4 bg-gray-200 rounded w-1/2"></div>
                        </div>
                    </div>
                ))}
            </div>
        );
    }

    if (error) {
        return (
            <div className="p-4 bg-red-50 border border-red-200 rounded-lg">
                <p className="text-red-600">{error}</p>
            </div>
        );
    }

    const getEventColor = (type: string) => {
        switch (type.toLowerCase()) {
            case 'financial':
                return 'border-green-500 bg-green-50';
            case 'legal':
                return 'border-blue-500 bg-blue-50';
            case 'political':
                return 'border-purple-500 bg-purple-50';
            case 'corruption':
                return 'border-red-500 bg-red-50';
            default:
                return 'border-gray-500 bg-gray-50';
        }
    };

    const getImportanceIndicator = (importance: number) => {
        const color = importance >= 0.7 ? 'bg-red-500' : 
                     importance >= 0.4 ? 'bg-yellow-500' : 
                     'bg-green-500';
        return (
            <span className={`inline-block w-2 h-2 rounded-full ${color} mr-2`} />
        );
    };

    return (
        <div className="relative">
            {/* Timeline line */}
            <div className="absolute left-24 top-0 bottom-0 w-0.5 bg-gray-200" />

            {/* Events */}
            <div className="space-y-8">
                {events.map((event) => (
                    <div
                        key={event.id}
                        className="relative flex items-start cursor-pointer group"
                        onClick={() => onEventClick?.(event)}
                    >
                        {/* Date */}
                        <div className="flex-none w-24 pt-2 text-sm text-gray-500">
                            {new Date(event.date).toLocaleDateString()}
                        </div>

                        {/* Dot */}
                        <div className="flex-none w-8">
                            <div className="w-3 h-3 rounded-full bg-white border-2 border-gray-300 group-hover:border-blue-500" />
                        </div>

                        {/* Content */}
                        <div className={`flex-grow p-4 rounded-lg border ${getEventColor(event.type)} transition-colors group-hover:border-blue-500`}>
                            <div className="flex items-center mb-2">
                                {getImportanceIndicator(event.importance)}
                                <h3 className="text-lg font-medium text-gray-900">{event.title}</h3>
                            </div>
                            
                            <p className="text-gray-600 mb-3">{event.description}</p>
                            
                            {/* Entity tags */}
                            <div className="flex flex-wrap gap-2">
                                {event.entities.map((entity, index) => (
                                    <span
                                        key={index}
                                        className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800"
                                    >
                                        {entity}
                                    </span>
                                ))}
                            </div>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
};

export default TimelineView;
