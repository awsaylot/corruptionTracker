import React, { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/router';

interface NavItem {
    label: string;
    href: string;
    icon?: string;
    description?: string;
    items?: {
        label: string;
        href: string;
        description?: string;
    }[];
}

const navigation: NavItem[] = [
    {
        label: 'Dashboard',
        href: '/',
        icon: '📊'
    },
    {
        label: 'Graph',
        href: '/graph',
        icon: '🕸️',
        items: [
            {
                label: 'Visualization',
                href: '/graph/visualizer',
                description: 'Interactive graph visualization'
            },
            {
                label: 'Search',
                href: '/graph/search',
                description: 'Advanced graph search and filtering'
            },
            {
                label: 'Operations',
                href: '/graph/operations',
                description: 'Graph operations and traversals'
            }
        ]
    },
    {
        label: 'Node Types',
        href: '/nodes',
        icon: '📝',
        items: [
            {
                label: 'Type Manager',
                href: '/nodes/types',
                description: 'Manage node types and properties'
            },
            {
                label: 'Create Node',
                href: '/nodes/create',
                description: 'Create new nodes'
            },
            {
                label: 'Batch Operations',
                href: '/nodes/batch',
                description: 'Bulk node operations'
            }
        ]
    },
    {
        label: 'Relationships',
        href: '/relationships',
        icon: '🔗',
        items: [
            {
                label: 'Type Manager',
                href: '/relationships/types',
                description: 'Manage relationship types'
            },
            {
                label: 'Create Relationship',
                href: '/relationships/create',
                description: 'Create new relationships'
            }
        ]
    },
    {
        label: 'Batch Operations',
        href: '/batch',
        icon: '📦',
        description: 'Batch data operations'
    },
    {
        label: 'Chat',
        href: '/chat',
        icon: '💬',
        description: 'AI Assistant Chat'
    },
    {
        label: 'Analytics',
        href: '/analytics',
        icon: '📈',
        items: [
            {
                label: 'Network Stats',
                href: '/analytics/network',
                description: 'Graph network statistics'
            },
            {
                label: 'Corruption Score',
                href: '/analytics/corruption',
                description: 'Entity corruption risk analysis'
            },
            {
                label: 'Timeline',
                href: '/analytics/timeline',
                description: 'Temporal analysis'
            }
        ]
    },
    {
        label: 'Import',
        href: '/import',
        icon: '📥',
        description: 'Import data into the system'
    },
    {
        label: 'Extraction',
        href: '/extraction',
        icon: '📝',
        description: 'Extract entities and relationships from text'
    }
];

const Navbar: React.FC = () => {
    const router = useRouter();
    const [activeDropdown, setActiveDropdown] = useState<string | null>(null);
    const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

    const isActive = (href: string) => router.pathname === href ||
        router.pathname.startsWith(href + '/');

    const handleMouseEnter = (label: string) => {
        setActiveDropdown(label);
    };

    const handleMouseLeave = () => {
        setActiveDropdown(null);
    };

    return (
        <nav className="bg-white shadow">
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                <div className="flex justify-between h-16">
                    <div className="flex">
                        <div className="flex-shrink-0 flex items-center">
                            <Link href="/" className="text-xl font-bold text-gray-800">
                                Graph Explorer
                            </Link>
                        </div>

                        <div className="hidden sm:ml-6 sm:flex sm:space-x-8">
                            {navigation.map((item) => (
                                <div
                                    key={item.label}
                                    className="relative"
                                    onMouseEnter={() => handleMouseEnter(item.label)}
                                    onMouseLeave={handleMouseLeave}
                                >
                                    <Link 
                                        href={item.href}
                                        className={`
                                            inline-flex items-center px-1 pt-1 border-b-2
                                            text-sm font-medium h-16
                                            ${isActive(item.href)
                                                ? 'border-blue-500 text-gray-900'
                                                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                                            }
                                        `}
                                    >
                                        <span className="mr-2">{item.icon}</span>
                                        {item.label}
                                    </Link>

                                    {item.items && activeDropdown === item.label && (
                                        <div className="absolute z-10 left-1/2 transform -translate-x-1/2 mt-0 px-2 w-screen max-w-md sm:px-0">
                                            <div className="rounded-lg shadow-lg ring-1 ring-black ring-opacity-5 overflow-hidden">
                                                <div className="relative grid gap-6 bg-white px-5 py-6 sm:gap-8 sm:p-8">
                                                    {item.items.map((subItem) => (
                                                        <Link
                                                            key={subItem.label}
                                                            href={subItem.href}
                                                            className="-m-3 p-3 flex items-start rounded-lg hover:bg-gray-50"
                                                        >
                                                            <div className="ml-4">
                                                                <p className="text-base font-medium text-gray-900">
                                                                    {subItem.label}
                                                                </p>
                                                                {subItem.description && (
                                                                    <p className="mt-1 text-sm text-gray-500">
                                                                        {subItem.description}
                                                                    </p>
                                                                )}
                                                            </div>
                                                        </Link>
                                                    ))}
                                                </div>
                                            </div>
                                        </div>
                                    )}
                                </div>
                            ))}
                        </div>
                    </div>

                    {/* Mobile menu button */}
                    <div className="flex items-center sm:hidden">
                        <button
                            type="button"
                            onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
                            className="inline-flex items-center justify-center p-2 rounded-md text-gray-400 hover:text-gray-500 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-blue-500"
                            aria-expanded="false"
                        >
                            <span className="sr-only">Open main menu</span>
                            <svg
                                className={`${isMobileMenuOpen ? 'hidden' : 'block'} h-6 w-6`}
                                xmlns="http://www.w3.org/2000/svg"
                                fill="none"
                                viewBox="0 0 24 24"
                                stroke="currentColor"
                                aria-hidden="true"
                            >
                                <path
                                    strokeLinecap="round"
                                    strokeLinejoin="round"
                                    strokeWidth={2}
                                    d="M4 6h16M4 12h16M4 18h16"
                                />
                            </svg>
                            <svg
                                className={`${isMobileMenuOpen ? 'block' : 'hidden'} h-6 w-6`}
                                xmlns="http://www.w3.org/2000/svg"
                                fill="none"
                                viewBox="0 0 24 24"
                                stroke="currentColor"
                                aria-hidden="true"
                            >
                                <path
                                    strokeLinecap="round"
                                    strokeLinejoin="round"
                                    strokeWidth={2}
                                    d="M6 18L18 6M6 6l12 12"
                                />
                            </svg>
                        </button>
                    </div>
                </div>
            </div>

            {/* Mobile menu */}
            <div className={`${isMobileMenuOpen ? 'block' : 'hidden'} sm:hidden`}>
                <div className="pt-2 pb-3 space-y-1">
                    {navigation.map((item) => (
                        <div key={item.label}>
                            <Link
                                href={item.href}
                                className={`
                                    block pl-3 pr-4 py-2 border-l-4 text-base font-medium
                                    ${isActive(item.href)
                                        ? 'bg-blue-50 border-blue-500 text-blue-700'
                                        : 'border-transparent text-gray-500 hover:bg-gray-50 hover:border-gray-300 hover:text-gray-700'
                                    }
                                `}
                            >
                                <span className="mr-2">{item.icon}</span>
                                {item.label}
                            </Link>
                            {item.items && (
                                <div className="ml-4 border-l border-gray-200">
                                    {item.items.map((subItem) => (
                                        <Link
                                            key={subItem.label}
                                            href={subItem.href}
                                            className="block pl-3 pr-4 py-2 text-base font-medium text-gray-500 hover:text-gray-700 hover:bg-gray-50"
                                        >
                                            {subItem.label}
                                        </Link>
                                    ))}
                                </div>
                            )}
                        </div>
                    ))}
                </div>
            </div>
        </nav>
    );
};

export default Navbar;
