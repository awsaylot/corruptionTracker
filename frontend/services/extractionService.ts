import { apiURL } from '../utils/constants';

export class ExtractionService {
  private baseUrl: string;

  constructor() {
    this.baseUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
  }

  async extractFromURL(url: string): Promise<any> {
    const response = await fetch(`${this.baseUrl}/extraction`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ url }),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(errorText || 'Failed to extract from URL');
    }

    return response.json();
  }

  async getAnalysisStatus(sessionId: string): Promise<any> {
    const response = await fetch(`${this.baseUrl}/extraction/${sessionId}/status`);
    
    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(errorText || 'Failed to get analysis status');
    }

    return response.json();
  }
}

export const extractionService = new ExtractionService();
