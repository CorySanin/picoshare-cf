"use strict";

async function uploadFormData(url, formData, progressFn) {
  const CHUNK_SIZE = 32768 * 10; // sqlite.go:defaultChunkSize
  const file = formData.get("file");
  const note = formData.get('note')
  const totalSize = file.size;
  const splits = Math.ceil(file.size / CHUNK_SIZE);
  let uploaded = 0;
  let id;

  for (const [i, _] of [...Array(splits)].entries()) {
    const start = i * CHUNK_SIZE;
    const chunk = file.slice(start, start + CHUNK_SIZE);
    const chunkData = new FormData();
    chunkData.append("file", chunk);
    chunkData.append("filename", file.name);
    chunkData.append("contentType", file.type);
    chunkData.append("last", i === splits - 1 ? "1" : "0");
    if (note) {
      chunkData.append("note", note);
    }
    if (id) {
      chunkData.append("id", id);
    }

    const response = await fetch(url, {
      method: "POST",
      body: chunkData,
      headers: { "Accept": "application/json" },
    });

    if (!response.ok) {
      throw new Error(`Upload failed: ${response.statusText}`);
    }

    uploaded += chunk.size;
    if (progressFn) {
      progressFn(uploaded, totalSize);
    }
    if (start > 0 && uploaded !== totalSize) {
      continue;
    }
    const data = await response.json();
    if (uploaded === totalSize) {
      return data;
    }
    if (!('id' in data) || data.id === '') {
      throw new Error("Missing expected id field");
    }
    id = data.id;
  }
}

function uploadFormDataXHR(url, formData, progressFn) {
  return new Promise((resolve, reject) => {
    // We have to use XHR instead of fetch because fetch currently doesn't
    // support a mechanism for reporting upload progress.
    const xhr = new XMLHttpRequest();
    xhr.open("POST", url, true);
    xhr.setRequestHeader("Accept", "application/json");
    xhr.upload.addEventListener("progress", (event) => {
      if (event.lengthComputable) {
        if (progressFn) {
          progressFn(event.loaded, event.total);
        }
      }
    });
    xhr.addEventListener("loadend", () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        resolve(xhr.response);
      } else {
        if (xhr.responseText) {
          reject(xhr.responseText);
          return;
        }
        reject(xhr.statusText);
      }
    });
    xhr.addEventListener("error", () => {
      reject(
        "Failed to communicate with server" +
          (xhr.statusText ? `: ${xhr.statusText}` : ".")
      );
    });
    xhr.send(formData);
  })
    .then((raw) => {
      return Promise.resolve(JSON.parse(raw));
    })
    .then((data) => {
      if (!Object.prototype.hasOwnProperty.call(data, "id")) {
        throw new Error("Missing expected id field");
      }
      return Promise.resolve(data);
    });
}

export async function uploadFile(file, expirationTime, note, progressFn) {
  const formData = new FormData();
  formData.append("file", file);
  if (note) {
    formData.append("note", note);
  }
  return uploadFormData(
    `/api/entry?expiration=${encodeURIComponent(expirationTime)}`,
    formData,
    progressFn
  );
}

export async function guestUploadFile(
  file,
  guestLinkID,
  expirationTime,
  progressFn
) {
  const formData = new FormData();
  formData.append("file", file);
  return uploadFormData(
    `/api/guest/${guestLinkID}?expiration=${encodeURIComponent(
      expirationTime
    )}`,
    formData,
    progressFn
  );
}

export async function editFile(id, filename, expiration, note) {
  let payload = {
    filename,
    note,
  };
  if (expiration) {
    payload.expiration = expiration;
  }
  return fetch(`/api/entry/${encodeURIComponent(id)}`, {
    method: "PUT",
    credentials: "include",
    body: JSON.stringify(payload),
  })
    .then((response) => {
      if (!response.ok) {
        return response.text().then((error) => {
          return Promise.reject(error);
        });
      }
      return Promise.resolve();
    })
    .catch((error) => {
      if (error.message) {
        return Promise.reject(
          "Failed to communicate with server" +
            (error.message ? `: ${error.message}` : ".")
        );
      }
      return Promise.reject(error);
    });
}

export async function deleteFile(id) {
  return fetch(`/api/entry/${id}`, {
    method: "DELETE",
    credentials: "include",
  })
    .then((response) => {
      if (!response.ok) {
        return response.text().then((error) => {
          return Promise.reject(error);
        });
      }
      return Promise.resolve();
    })
    .catch((error) => {
      if (error.message) {
        return Promise.reject(
          "Failed to communicate with server" +
            (error.message ? `: ${error.message}` : ".")
        );
      }
      return Promise.reject(error);
    });
}
