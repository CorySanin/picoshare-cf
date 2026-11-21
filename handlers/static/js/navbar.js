import { logOut } from "./controllers/auth.js";

const logOutEl = document.getElementById("navbar-log-out");
if (logOutEl) {
  logOutEl.addEventListener("click", () => {
    logOut().then(() => {
      document.location = "/";
    });
  });
}

const uploadNavEl = document.getElementById("navbar-upload-link");
if (uploadNavEl) {
  uploadNavEl.addEventListener("click", (evt) => {
    const fileInput = document.querySelector(".file-input");
    if (!fileInput) {
      // Outside the upload page, keep the default link navigation.
      return;
    }
    evt.preventDefault();
    fileInput.click();
  });
}
