import io, os
import numpy as np
from fastapi import FastAPI, UploadFile, Request
from PIL import Image
from sentence_transformers import SentenceTransformer
import torch

device = "mps" if torch.backends.mps.is_available() else "cpu"
model = SentenceTransformer(os.getenv("CLIP_MODEL", "clip-ViT-B-32"), device=device)
app = FastAPI()

def _norm(v: np.ndarray) -> list[float]:
    return (v / np.linalg.norm(v)).astype(float).tolist()  # L2-normalize for cosine

@app.post("/embed/text")
async def embed_text(req: Request):
    body = await req.json()
    return {"vector": _norm(model.encode(body["text"], convert_to_numpy=True))}

@app.post("/embed/image")
async def embed_image(file: UploadFile):
    img = Image.open(io.BytesIO(await file.read())).convert("RGB")
    return {"vector": _norm(model.encode(img, convert_to_numpy=True))}

@app.get("/health")
def health():
    return {"status": "ok", "device": device,
            "dim": model.get_sentence_embedding_dimension()}
