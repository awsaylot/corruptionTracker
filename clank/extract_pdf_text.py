import PyPDF2
import re
import sys

def extract_text_from_pdf(pdf_path):
    text = ""
    try:
        with open(pdf_path, 'rb') as file:
            pdf_reader = PyPDF2.PdfReader(file)
            for page in pdf_reader.pages:
                text += page.extract_text() + "\n"
    except Exception as e:
        print(f"Error reading {pdf_path}: {e}")
        return None
    return text

def extract_names_and_relationships(text):
    # Look for patterns that indicate names and relationships
    names = set()
    relationships = []
    
    # Common name patterns - look for capitalized words that could be names
    name_pattern = r'\b[A-Z][a-z]+(?:\s+[A-Z][a-z]+)*\b'
    potential_names = re.findall(name_pattern, text)
    
    # Filter out common words that aren't names
    common_words = {'The', 'This', 'That', 'They', 'There', 'Then', 'These', 'Those', 
                   'When', 'Where', 'What', 'Why', 'How', 'Who', 'Which', 'Interview',
                   'Transcript', 'Maxwell', 'Redacted', 'Page', 'Question', 'Answer',
                   'Detective', 'Officer', 'Mr', 'Mrs', 'Ms', 'Dr', 'Professor'}
    
    for name in potential_names:
        if name not in common_words and len(name) > 2:
            names.add(name)
    
    # Look for relationship indicators
    relationship_patterns = [
        r'([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*)\s+(?:is|was)\s+(?:my|his|her|their)\s+([a-z]+)',
        r'([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*)\s+and\s+([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*)\s+are\s+([a-z]+)',
        r'([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*)\s+works?\s+(?:with|for)\s+([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*)',
        r'([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*)\s+knows?\s+([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*)',
        r'([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*)\s+(?:met|meets?)\s+([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*)',
    ]
    
    for pattern in relationship_patterns:
        matches = re.findall(pattern, text, re.IGNORECASE)
        relationships.extend(matches)
    
    # Look for REDACTED relationships
    redacted_patterns = [
        r'REDACTED\s+(?:is|was)\s+(?:my|his|her|their)\s+([a-z]+)',
        r'REDACTED\s+and\s+([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*)\s+are\s+([a-z]+)',
        r'([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*)\s+and\s+REDACTED\s+are\s+([a-z]+)',
        r'REDACTED\s+works?\s+(?:with|for)\s+([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*)',
        r'([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*)\s+works?\s+(?:with|for)\s+REDACTED',
    ]
    
    for pattern in redacted_patterns:
        matches = re.findall(pattern, text, re.IGNORECASE)
        relationships.extend([('REDACTED', match) if isinstance(match, str) else match for match in matches])
    
    return names, relationships

if __name__ == "__main__":
    pdf1 = r"C:\Users\Celot\Downloads\Interview Transcript - Maxwell 2025.07.24 (Redacted).pdf"
    pdf2 = r"C:\Users\Celot\Downloads\Interview Transcript - Maxwell 2025.07.25-cft (Redacted).pdf"
    
    print("=== ANALYZING FIRST PDF ===")
    text1 = extract_text_from_pdf(pdf1)
    if text1:
        names1, relationships1 = extract_names_and_relationships(text1)
        print("Names found:")
        for name in sorted(names1):
            print(f"  - {name}")
        print("\nRelationships found:")
        for rel in relationships1:
            print(f"  - {rel}")
        
        # Save first part of text for manual review
        with open("pdf1_sample.txt", "w", encoding="utf-8") as f:
            f.write(text1[:5000])
    
    print("\n=== ANALYZING SECOND PDF ===")
    text2 = extract_text_from_pdf(pdf2)
    if text2:
        names2, relationships2 = extract_names_and_relationships(text2)
        print("Names found:")
        for name in sorted(names2):
            print(f"  - {name}")
        print("\nRelationships found:")
        for rel in relationships2:
            print(f"  - {rel}")
            
        # Save first part of text for manual review
        with open("pdf2_sample.txt", "w", encoding="utf-8") as f:
            f.write(text2[:5000])
