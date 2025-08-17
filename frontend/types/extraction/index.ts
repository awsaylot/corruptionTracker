// frontend/types/extraction.ts
export interface Article {
  id: string;
  url: string;
  title: string;
  content: string;
  source: string;
  author: string;
  publishDate: string;
  extractedAt: string;
  metadata: Record<string, any>;
}

export interface EntityMention {
  text: string;
  context: string;
  position?: {
    start: number;
    end: number;
  };
}

export interface ExtractedEntity {
  id: string;
  type: 'person' | 'organization' | 'location' | 'money' | 'time';
  name: string;
  properties: Record<string, any>;
  confidence: number;
  mentions: EntityMention[];
  articleId: string;
  extractedAt: string;
}

export interface ExtractedRelationship {
  id: string;
  type: string;
  fromId: string;
  toId: string;
  properties: Record<string, any>;
  confidence: number;
  context: string;
  articleId: string;
  extractedAt: string;
}

export interface ExtractionResult {
  entities: ExtractedEntity[];
  relationships: ExtractedRelationship[];
  confidence: number;
}

export interface AnalysisConfig {
  depth: number;
  maxStages: number;
  confidenceThreshold: number;
  timeoutPerStage: number;
  enableCrossReference: boolean;
  enableHypotheses: boolean;
}

export interface AnalysisStage {
  id: string;
  name: string;
  description: string;
  stage: number;
  status: 'pending' | 'running' | 'completed' | 'failed';
  startedAt: string;
  completedAt?: string;
  results?: ExtractionResult;
  confidence: number;
  insights: string[];
  questions: string[];
  error?: string;
}

export interface EvidenceChain {
  id: string;
  claim: string;
  evidence: string[];
  confidence: number;
  sources: string[];
  createdAt: string;
}

export interface Hypothesis {
  id: string;
  description: string;
  confidence: number;
  supporting: string[];
  conflicting: string[];
  questions: string[];
  createdAt: string;
}

export interface AnalysisSession {
  id: string;
  articleId: string;
  config: AnalysisConfig;
  stages: AnalysisStage[];
  currentStage: number;
  status: 'running' | 'completed' | 'failed' | 'terminated';
  startedAt: string;
  completedAt?: string;
  finalResults?: ExtractionResult;
  evidence: EvidenceChain[];
  hypotheses: Hypothesis[];
}

export interface ExtractionRequest {
  url: string;
  depth?: number;
}

export interface ExtractionResponse {
  sessionId: string;
  article: Article;
  session: AnalysisSession;
  status: string;
}

// Processing states
export type ProcessingState = 
  | 'idle'
  | 'validating'
  | 'scraping'
  | 'processing'
  | 'analyzing'
  | 'completed'
  | 'error';

export interface ProcessingStatus {
  state: ProcessingState;
  message: string;
  progress?: number;
  error?: string;
}

// Analysis depth configuration
export interface DepthConfig {
  level: number;
  name: string;
  description: string;
  estimatedTime: string;
  features: string[];
}

export const DEPTH_CONFIGS: DepthConfig[] = [
  {
    level: 2,
    name: 'Basic',
    description: 'Surface-level entity extraction and basic relationships',
    estimatedTime: '30-60 seconds',
    features: ['Entity identification', 'Basic relationships', 'Confidence scoring']
  },
  {
    level: 3,
    name: 'Standard',
    description: 'Enhanced analysis with pattern recognition',
    estimatedTime: '1-2 minutes',
    features: ['Deep entity analysis', 'Pattern recognition', 'Cross-referencing']
  },
  {
    level: 5,
    name: 'Advanced',
    description: 'Comprehensive analysis with corruption indicators',
    estimatedTime: '2-3 minutes',
    features: ['Corruption patterns', 'Evidence chains', 'Risk assessment']
  },
  {
    level: 7,
    name: 'Deep',
    description: 'Full investigative analysis with hypotheses',
    estimatedTime: '3-5 minutes',
    features: ['Hypothesis generation', 'Investigation leads', 'Gap analysis']
  },
  {
    level: 10,
    name: 'Comprehensive',
    description: 'Maximum depth with recursive refinement',
    estimatedTime: '5-10 minutes',
    features: ['Recursive analysis', 'Multi-stage validation', 'Complete synthesis']
  }
];

// Validation errors
export interface ValidationError {
  field: string;
  message: string;
  code: string;
}

// API response helpers
export interface APIResponse<T> {
  data?: T;
  error?: string;
  status: number;
}

// Progress tracking
export interface ProgressUpdate {
  sessionId: string;
  stage: number;
  stageName: string;
  progress: number;
  message: string;
  confidence?: number;
  insights?: string[];
}

// Analysis metrics
export interface AnalysisMetrics {
  entitiesExtracted: number;
  relationshipsFound: number;
  averageConfidence: number;
  processingTime: number;
  stagesCompleted: number;
  hypothesesGenerated: number;
  evidenceChainsFound: number;
}

// frontend/types/analysis.ts
export interface CorruptionIndicators {
  financial_irregularities: string[];
  procedural_violations: string[];
  behavioral_red_flags: string[];
  conflict_of_interest: string[];
  abuse_of_power: string[];
  lack_of_transparency: string[];
}

export interface ConfidenceAssessment {
  overall_confidence: number;
  data_quality: 'high' | 'medium' | 'low';
  evidence_strength: 'strong' | 'moderate' | 'weak';
}

export interface InvestigativeLead {
  lead_type: 'document_trail' | 'witness' | 'financial_analysis' | 'pattern_analysis';
  description: string;
  priority: 'high' | 'medium' | 'low';
  estimated_effort: 'low' | 'medium' | 'high';
  potential_impact: 'breakthrough' | 'supporting' | 'minor';
}

export interface LegalImplications {
  potential_charges: string[];
  jurisdiction: 'federal' | 'state' | 'local';
  statute_of_limitations: 'within' | 'expired' | 'unclear';
  prosecution_likelihood: 'high' | 'medium' | 'low';
}

export interface RiskAssessment {
  ongoing_corruption_risk: 'high' | 'medium' | 'low';
  reputational_damage_risk: 'high' | 'medium' | 'low';
  financial_loss_risk: 'high' | 'medium' | 'low';
  systemic_corruption_indicators: 'yes' | 'no' | 'unclear';
}

export interface FinalAnalysis {
  executive_summary: {
    corruption_likelihood: number;
    primary_corruption_types: string[];
    key_actors: Array<{
      name: string;
      role: string;
      involvement_level: 'high' | 'medium' | 'low';
    }>;
    financial_scope: {
      estimated_amount: string;
      currency: string;
      certainty: 'high' | 'medium' | 'low';
    };
    institutional_impact: 'high' | 'medium' | 'low';
    timeline_span: string;
  };
  final_entities: ExtractedEntity[];
  final_relationships: ExtractedRelationship[];
  corruption_indicators: CorruptionIndicators;
  evidence_strength_assessment: {
    documentary_evidence: 'strong' | 'moderate' | 'weak';
    witness_testimony: 'strong' | 'moderate' | 'weak';
    financial_evidence: 'strong' | 'moderate' | 'weak';
    circumstantial_evidence: 'strong' | 'moderate' | 'weak';
    overall_evidence_quality: 'strong' | 'moderate' | 'weak';
  };
  investigative_priorities: Array<{
    priority: number;
    action: string;
    rationale: string;
    difficulty: 'easy' | 'medium' | 'hard';
    potential_impact: 'high' | 'medium' | 'low';
  }>;
  legal_implications: LegalImplications;
  risk_assessment: RiskAssessment;
  next_steps: string[];
  confidence_final: number;
  analysis_limitations: string[];
}

// Stage-specific analysis types
export interface StageConfiguration {
  name: string;
  enabled: boolean;
  timeout: number;
  confidenceThreshold: number;
  parameters: Record<string, any>;
}

export interface ValidationResults {
  consistency_score: number;
  fact_verification_score: number;
  timeline_coherence_score: number;
  logical_coherence_score: number;
  evidence_strength_score: number;
  overall_validation_score: number;
}

export interface ValidationIssue {
  type: 'inconsistency' | 'unsupported_claim' | 'timeline_error' | 'logical_error' | 'weak_evidence';
  severity: 'high' | 'medium' | 'low';
  description: string;
  affected_entities: string[];
  affected_relationships: string[];
  suggested_correction: string;
  confidence_impact: string;
}

// Analysis progress tracking
export interface StageProgress {
  stage_number: number;
  stage_name: string;
  status: 'pending' | 'running' | 'completed' | 'failed';
  progress_percentage: number;
  estimated_completion: string;
  current_task: string;
}

export interface AnalysisProgress {
  session_id: string;
  overall_progress: number;
  current_stage: number;
  total_stages: number;
  stages: StageProgress[];
  estimated_completion: string;
  can_terminate: boolean;
}

// Evidence and hypothesis types
export interface EvidenceSource {
  type: 'article' | 'document' | 'testimony' | 'financial_record';
  description: string;
  reliability: 'high' | 'medium' | 'low';
  source_date: string;
}

export interface EnhancedEvidenceChain extends EvidenceChain {
  sources_detail: EvidenceSource[];
  related_entities: string[];
  corroboration_level: 'strong' | 'moderate' | 'weak';
  legal_admissibility: 'admissible' | 'questionable' | 'inadmissible';
}

export interface EnhancedHypothesis extends Hypothesis {
  type: 'corruption' | 'fraud' | 'bribery' | 'embezzlement' | 'conflict_of_interest' | 'cover_up';
  required_evidence: string[];
  implications: string[];
  investigation_priority: 'high' | 'medium' | 'low';
  testability: 'high' | 'medium' | 'low';
  related_hypotheses: string[];
}

// Analysis export and sharing
export interface AnalysisExport {
  session_id: string;
  export_format: 'json' | 'pdf' | 'csv' | 'xlsx';
  include_raw_data: boolean;
  include_hypotheses: boolean;
  include_evidence: boolean;
  confidentiality_level: 'public' | 'internal' | 'confidential';
}

// Real-time updates
export interface AnalysisUpdate {
  type: 'stage_start' | 'stage_complete' | 'progress' | 'insight' | 'error' | 'complete';
  session_id: string;
  timestamp: string;
  data: any;
}

// Analysis comparison (for multiple articles)
export interface AnalysisComparison {
  sessions: string[];
  common_entities: ExtractedEntity[];
  common_patterns: string[];
  divergent_findings: string[];
  correlation_score: number;
  synthesis_insights: string[];
}

// Error handling
export interface AnalysisError {
  code: string;
  message: string;
  stage?: string;
  recoverable: boolean;
  suggested_action: string;
  technical_details?: string;
}

export const ERROR_CODES = {
  INVALID_URL: 'INVALID_URL',
  SCRAPING_FAILED: 'SCRAPING_FAILED',
  CONTENT_TOO_SHORT: 'CONTENT_TOO_SHORT',
  LLM_TIMEOUT: 'LLM_TIMEOUT',
  PARSING_ERROR: 'PARSING_ERROR',
  CONFIDENCE_TOO_LOW: 'CONFIDENCE_TOO_LOW',
  STAGE_TIMEOUT: 'STAGE_TIMEOUT',
  SESSION_NOT_FOUND: 'SESSION_NOT_FOUND',
  ANALYSIS_TERMINATED: 'ANALYSIS_TERMINATED',
} as const;

// Utility types
export type EntityType = ExtractedEntity['type'];
export type AnalysisStatus = AnalysisSession['status'];
export type StageStatus = AnalysisStage['status'];
export type ConfidenceLevel = 'high' | 'medium' | 'low';
export type AnalysisDepth = 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10;

// Helper functions for types
export const getConfidenceLevel = (confidence: number): ConfidenceLevel => {
  if (confidence >= 0.8) return 'high';
  if (confidence >= 0.6) return 'medium';
  return 'low';
};

export const getDepthConfig = (depth: number): DepthConfig | undefined => {
  return DEPTH_CONFIGS.find(config => config.level === depth) || 
         DEPTH_CONFIGS.find(config => config.level <= depth)?.slice(-1)[0];
};

export const formatConfidence = (confidence: number): string => {
  return `${(confidence * 100).toFixed(1)}%`;
};

export const getStatusColor = (status: StageStatus | AnalysisStatus): string => {
  switch (status) {
    case 'completed':
      return 'text-green-600 bg-green-100';
    case 'running':
      return 'text-blue-600 bg-blue-100';
    case 'failed':
      return 'text-red-600 bg-red-100';
    case 'terminated':
      return 'text-gray-600 bg-gray-100';
    default:
      return 'text-gray-600 bg-gray-100';
  }
};

// Constants
export const MAX_ANALYSIS_DEPTH = 10;
export const MIN_ANALYSIS_DEPTH = 2;
export const DEFAULT_ANALYSIS_DEPTH = 3;
export const DEFAULT_CONFIDENCE_THRESHOLD = 0.6;
export const DEFAULT_STAGE_TIMEOUT = 60000; // 60 seconds

export const SUPPORTED_NEWS_DOMAINS = [
  'reuters.com',
  'ap.org', 
  'bbc.com',
  'cnn.com',
  'nytimes.com',
  'washingtonpost.com',
  'guardian.co.uk',
  'ft.com',
  'wsj.com',
  'bloomberg.com',
  'politico.com',
  'propublica.org',
  'axios.com',
  'npr.org',
  'abcnews.go.com',
  'cbsnews.com',
  'nbcnews.com',
  'usatoday.com',
  'latimes.com'
] as const;

export type SupportedNewsDomain = typeof SUPPORTED_NEWS_DOMAINS[number];