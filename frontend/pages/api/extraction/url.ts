import type { NextApiRequest, NextApiResponse } from 'next';
import { extractionService } from '../../../services/extractionService';

export default async function handler(req: NextApiRequest, res: NextApiResponse) {
  if (req.method !== 'POST') {
    return res.status(405).json({ error: 'Method not allowed' });
  }

  try {
    const { url } = req.body;
    console.log('[API] Forwarding extraction request to Clank:', { url });
    
    const data = await extractionService.extractFromURL(url);
    console.log('[API] Extraction response received:', data);
    
    res.status(200).json(data);
  } catch (error) {
    console.error('[API] Extraction error:', error);
    res.status(500).json({ 
      error: error instanceof Error ? error.message : 'Failed to process extraction request' 
    });
  }
}
