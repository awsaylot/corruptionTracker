interface Relationship {
  source: string;
  target: string;
  type: string;
}

export interface ExtractedContent {
  url: string;
  title: string;
  summary: string;
  entities: string[];
  relationships: Relationship[];
  raw?: string;
}

export interface ExtractionResponse {
  success: boolean;
  data: ExtractedContent;
  error?: string;
}
