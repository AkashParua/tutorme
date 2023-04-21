import PyPDF2
from flask import Flask, request 
import os 
import http.client
import pinecone
from dotenv import load_dotenv  
import json
from flask import jsonify
from sentence_transformers import SentenceTransformer
model = SentenceTransformer("all-MiniLM-L6-v2")
load_dotenv()
pinecone.init(api_key= os.getenv('api_key'), environment=os.getenv('type'))
index = pinecone.Index('text-chunks')

app = Flask(__name__)


@app.route('/upload', methods=['POST'])
def upload_file():
    file = request.files['file']
    filename = file.filename
    maxsize = 2000
    file.save(os.path.join(os.path.abspath('pdfs'), filename))
    with open(os.path.join(os.path.abspath('pdfs'), filename), 'rb') as f:
        pdf_reader = PyPDF2.PdfReader(f)
        text = ""
        for page_num in range(len(pdf_reader.pages)):
            page = pdf_reader.pages[page_num]
            text += page.extract_text()
        lines = text.split('\n')
        text_chunks = []
        text_chunk = ''
        for line in lines:
            if len(line) + len(text_chunk) > maxsize:
                text_chunk.strip()
                text_chunks.append(text_chunk)
                text_chunk = ''
            else :
                text_chunk += line
                              
    for i, chunk in enumerate(text_chunks):
        # and store them in Pinecone
        chunkInfo = (str(filename+'-'+str(i)),model.encode(chunk).tolist(),{'Book': filename, 'context':chunk })
        index.upsert(vectors=[chunkInfo])
    return 'File uploaded successfully'


@app.route('/delete' , methods = ['DELETE'])
def delete_file():
    filename = request.args.get('filename')
    query_response = index.delete(filter={'Book' : filename})
    
    
    
    

@app.route('/search', methods=['GET'])
def search():
    question = request.args.get('question')
    query_em = model.encode(question).tolist()
    results = index.query(query_em, top_k=1 , include_metadata=True)
    print(results)
    result = {}
    result['id'] = results['matches'][0]['id']
    result['context'] = results['matches'][0]['metadata']['context']
    json_string = json.dumps(result)
    return jsonify(json_string)
    
@app.route('/search_and_ask' , methods= ['GET'])
def searchandask():
    question = request.args.get('question')
    query_em = model.encode(question).tolist()
    results = index.query(query_em, top_k=1 , include_metadata=True)
    print(results)
    context = results['matches'][0]['metadata']['context']
    conn = http.client.HTTPSConnection("chatgpt-gpt-3-5.p.rapidapi.com")
     
    payload =  json.dumps({"query" : "Given the following context :"+context+question })
                                                                   
    headers = {
        'content-type': "application/json",
        'X-RapidAPI-Key': "2cc2e51559msh8bb526082cca664p1ff7cejsn0005b8d416ef",
        'X-RapidAPI-Host': "chatgpt-gpt-3-5.p.rapidapi.com"
        }

    conn.request("POST", "/ask", payload, headers)
    res = conn.getresponse()
    data = res.read()
    return json.loads(data)

    
    

if __name__ == '__main__':
  
    app.run()