import i18next from "i18next";
import { initReactI18next } from "react-i18next";

import { translationEn } from "./en";
import { translationRu } from "./ru";

const i18nextInstance = i18next.createInstance({
  resources: {
    en: { translation: translationEn },
    ru: { translation: translationRu }
  },
  lng: "en",
  fallbackLng: "en",
  interpolation: { escapeValue: false }
});

i18nextInstance.use(initReactI18next).init();

export default i18nextInstance;
