#!/usr/bin/env node
// Gera o ADMIN_PASSWORD_HASH (formato scrypt$<salt>$<hash>) a partir de uma
// senha, para o login do admin do blog. A senha em si nunca é guardada.
//
//   node scripts/hash-password.mjs 'minha-senha-forte'
//
// Cole a saída em ADMIN_PASSWORD_HASH (.env local) ou sele no cluster (kubeseal).
import { randomBytes, scryptSync } from "node:crypto";

const password = process.argv[2];
if (!password) {
  console.error("uso: node scripts/hash-password.mjs '<senha>'");
  process.exit(1);
}

const salt = randomBytes(16);
const hash = scryptSync(password, salt, 32);
console.log(`scrypt$${salt.toString("hex")}$${hash.toString("hex")}`);
