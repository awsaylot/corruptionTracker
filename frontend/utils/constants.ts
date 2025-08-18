// API URLs
export const apiURL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// API endpoints for llama.cpp server
export const API_ENDPOINTS = {
  CHAT: '/v1/chat/completions',
  STATUS: '/status',   // llama.cpp doesn't have this, but you could implement one
  HEALTH: '/health',   // same here — optional if you build one
} as const;

export const UI_CONSTANTS = {
  MAX_INPUT_LENGTH: 2048,
  INTERFACE_VERSION: 'v2.1',
  MODEL_NAME: 'CodeLlama-7b-Instruct-hf.Q6_K.gguf',
  UPDATE_INTERVAL: 1000,
  CONNECTION_TIMEOUT: 2000,
} as const;

export const SYSTEM_MESSAGES = {
  INIT: 'NEURAL INTERFACE v2.1 INITIALIZED...',
  READY: 'DEEPSEEK NEURAL NETWORK STATUS: READY',
  AWAITING: 'AWAITING USER INPUT...',
  CONNECTING: 'ESTABLISHING NEURAL LINK...',
  ERROR: 'NEURAL PATHWAY ERROR DETECTED',
} as const;

export const STATUS_MESSAGES = {
  QUANTUM_CORE: '> QUANTUM AI CORE: ONLINE',
  NEURAL_PATHWAYS: '> NEURAL PATHWAYS: ESTABLISHED',
  SECURITY: '> SECURITY PROTOCOLS: ACTIVE',
  ENCRYPTION: '> ENCRYPTION LAYER: ENABLED',
} as const;

export const THEME = {
  PRIMARY: '#00ff41',
  BACKGROUND: '#000000',
  ACCENT: '#001100',
} as const;
