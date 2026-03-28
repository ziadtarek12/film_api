import numpy as np
import json
import os
from neo4j import GraphDatabase

# Neo4j Connection config
URI = "bolt://localhost:7687"
AUTH = ("neo4j", "password") # Update with your credentials

def main():
    # 1. Load Data with Memory Mapping
    print("Loading data...")
    # mmap_mode checks out the file without filling up your system memory
    embeddings = np.load('imdb_embeddings_bge_m3.npy', mmap_mode='r')
    ids = np.load('movie_ids.npy')
    
    total_records = len(ids)
    print(f"Verification: {total_records} IDs and {embeddings.shape[0]} Vectors.")
    
    # 2. Checkpointing config
    checkpoint_file = 'ingest_checkpoint.json'
    start_idx = 0
    if os.path.exists(checkpoint_file):
        with open(checkpoint_file, 'r') as f:
            start_idx = json.load(f).get('last_processed_idx', 0)
        print(f"Checkpoint found. Resuming from record {start_idx}...")
        
    batch_size = 3000          # The sweet spot! (2000 - 5000)
    checkpoint_interval = 50000 
    
    driver = GraphDatabase.driver(URI, auth=AUTH)
    
    # The Cypher query used to insert the embedding back into Neo4j
    query = """
    UNWIND $batch AS row
    MATCH (f:Film {imdb_id: row.id})
    CALL db.create.setNodeVectorProperty(f, 'embedding', row.embedding)
    YIELD node
    RETURN count(*)
    """
    
    try:
        with driver.session() as session:
            for i in range(start_idx, total_records, batch_size):
                end_idx = min(i + batch_size, total_records)
                
                # Convert the numpy values back to standard Python lists for the neo4j driver
                batch_data = [
                    {
                        "id": str(ids[j]), 
                        "embedding": embeddings[j].tolist() 
                    } 
                    for j in range(i, end_idx)
                ]
                
                # Run batch transaction
                session.run(query, batch=batch_data)
                
                # Save checkpoint 
                if end_idx % checkpoint_interval <= batch_size or end_idx == total_records:
                    print(f"Checkpoint: Processed {end_idx}/{total_records} embeddings.")
                    with open(checkpoint_file, 'w') as f:
                        json.dump({'last_processed_idx': end_idx}, f)
                        
    except Exception as e:
        print(f"Error during ingestion: {str(e)}")
    finally:
        driver.close()
        print("Done!")

if __name__ == "__main__":
    main()
