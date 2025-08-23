import os
import re

# First, let's try to read the PDFs as binary and look for text patterns
def analyze_pdf_binary(file_path):
    print(f"Analyzing {file_path}")
    try:
        with open(file_path, 'rb') as f:
            # Read first 50KB to get a sample
            content = f.read(50000).decode('utf-8', errors='ignore')
            
            # Look for text patterns in the binary
            text_parts = re.findall(r'[A-Za-z\s]{10,}', content)
            readable_text = ' '.join(text_parts)
            
            print(f"Sample text found:")
            print(readable_text[:2000])
            print("=" * 50)
            
            # Extract potential names (capitalized words)
            names = set()
            name_pattern = r'\b[A-Z][a-z]{2,}(?:\s+[A-Z][a-z]{2,})*\b'
            potential_names = re.findall(name_pattern, readable_text)
            
            common_words = {'The', 'This', 'That', 'They', 'There', 'Then', 'These', 'Those', 
                           'When', 'Where', 'What', 'Why', 'How', 'Who', 'Which', 'Interview',
                           'Transcript', 'Maxwell', 'Redacted', 'Page', 'Question', 'Answer',
                           'Detective', 'Officer', 'And', 'But', 'For', 'Not', 'You', 'Are',
                           'Was', 'Were', 'Been', 'Have', 'Has', 'Had', 'Will', 'Would',
                           'Could', 'Should', 'May', 'Might', 'Can', 'Must', 'Shall'}
            
            for name in potential_names:
                if name not in common_words and len(name) > 2:
                    names.add(name)
            
            print(f"Potential names extracted:")
            for name in sorted(names)[:20]:  # Show first 20
                print(f"  - {name}")
            
            # Look for REDACTED mentions
            redacted_count = len(re.findall(r'REDACTED', readable_text, re.IGNORECASE))
            print(f"REDACTED mentions found: {redacted_count}")
            
            # Look for relationship words
            relationship_words = ['partner', 'colleague', 'friend', 'associate', 'boss', 'employee', 
                                'supervisor', 'manager', 'director', 'works with', 'knows', 'met']
            
            relationships_found = []
            for word in relationship_words:
                if word.lower() in readable_text.lower():
                    relationships_found.append(word)
            
            print(f"Relationship terms found: {relationships_found}")
            
            return readable_text, names
            
    except Exception as e:
        print(f"Error reading {file_path}: {e}")
        return None, set()

# Analyze both PDFs
pdf1_path = r"C:\Users\Celot\Downloads\Interview Transcript - Maxwell 2025.07.24 (Redacted).pdf"
pdf2_path = r"C:\Users\Celot\Downloads\Interview Transcript - Maxwell 2025.07.25-cft (Redacted).pdf"

print("=== FIRST PDF ANALYSIS ===")
text1, names1 = analyze_pdf_binary(pdf1_path)

print("\n=== SECOND PDF ANALYSIS ===")
text2, names2 = analyze_pdf_binary(pdf2_path)

print(f"\n=== SUMMARY ===")
all_names = names1.union(names2)
print(f"Total unique names found: {len(all_names)}")
print("All names:")
for name in sorted(all_names):
    print(f"  - {name}")
