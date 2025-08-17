package prompts

const ArticleExtractionPrompt = `Extract entities and relationships from the following news article. Focus on people, organizations, locations, and events related to corruption or misconduct.

Article:
{{.Content}}

Please analyze the article and provide the following information in JSON format:

{
  "entities": [
    {
      "type": "PERSON | ORGANIZATION | LOCATION",
      "name": "string",
      "properties": {
        "role": "string",
        "age": "number (optional)",
        "nationality": "string (optional)",
        "sector": "string (for organizations)",
        // Additional relevant properties
      },
      "confidence": "number (0-1)",
      "mentions": [
        {
          "text": "exact text from article",
          "context": "surrounding sentence or paragraph"
        }
      ]
    }
  ],
  "relationships": [
    {
      "type": "INVOLVED_IN | ACCUSED_OF | INVESTIGATED_FOR | CONNECTED_TO",
      "from": "entity name",
      "to": "entity name",
      "properties": {
        "date": "string (if available)",
        "amount": "number (if money involved)",
        "details": "string",
        // Additional relevant properties
      },
      "confidence": "number (0-1)",
      "context": "relevant quote or paragraph from article"
    }
  ]
}

Additional Instructions:
1. Focus on corruption-related relationships and roles
2. Include relevant monetary amounts and dates
3. Note any official investigations or accusations
4. Capture organizational hierarchies and connections
5. Maintain factual accuracy, avoid speculation`

const EntityValidationPrompt = `Validate the following extracted entities for accuracy and completeness:

Entities:
{{.Entities}}

Original Text:
{{.Content}}

For each entity:
1. Verify name accuracy
2. Confirm role and properties
3. Check for missing important information
4. Evaluate confidence score
5. Suggest any corrections

Provide response in JSON format:
{
  "validations": [
    {
      "entityId": "string",
      "isValid": boolean,
      "confidence": number,
      "suggestedCorrections": {
        // properties to correct
      },
      "missingInfo": [
        // list of missing important properties
      ]
    }
  ]
}`
