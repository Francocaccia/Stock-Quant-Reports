import { initializeApp } from "firebase/app";
import { getAuth } from "firebase/auth";

const firebaseConfig = {
  apiKey: "AIzaSyB-FNdCD0DoUcrReDfRCpiWnzEZh8ByIqk",
  authDomain: "quant-system-lynchburg-smif.firebaseapp.com",
  projectId: "quant-system-lynchburg-smif",
  storageBucket: "quant-system-lynchburg-smif.firebasestorage.app",
  messagingSenderId: "280156356401",
  appId: "1:280156356401:web:8222098f494ae852367a10"
};

// Initialize Firebase
const app = initializeApp(firebaseConfig);
export const auth = getAuth(app);
