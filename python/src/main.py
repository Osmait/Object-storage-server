import http
from fastapi import FastAPI, File, UploadFile, HTTPException
from fastapi.responses import JSONResponse
from pydantic import BaseModel
import os


app = FastAPI()


class Bucket(BaseModel):
    name: str


@app.post("/uploaded")
async def upload_file(file: UploadFile = File(...)):
    try:
        contents = await file.read()

        if not file.filename:
            raise HTTPException(status_code=400, detail="No filename provided ")

        file_path = os.path.join("uploads", file.filename)

        with open(file_path, "wb") as f:
            f.write(contents)

        return JSONResponse(
            {"message": f"File uploaded successfully: {file.filename}"},
            status_code=http.HTTPStatus.CREATED,
        )

    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/create")
async def create_bucket(bucket: Bucket) -> JSONResponse:
    try:
        os.makedirs(bucket.name, exist_ok=True)
        return JSONResponse(
            {"message": f"Bucket created successfully: {bucket.name}"},
            status_code=http.HTTPStatus.OK,
        )
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/list")
async def get_files(name: str):
    if not name:
        raise HTTPException(status_code=400, detail="name is required")

    if not os.path.exists(name) or not os.path.isdir(name):
        raise HTTPException(status_code=400, detail="Directory does not exist")

    files = [os.path.join(name, f) for f in os.listdir(name)]
    return JSONResponse(content=files)


@app.get("/")
async def hello_world():
    return {"message": "hello world"}

