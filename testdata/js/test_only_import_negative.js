// Test fixture: normal imports in production source
import { useState } from 'react';
import { db } from './database';
import { utils } from '../lib/utils';
import { config } from './config';

export function getUser() {
  return db.query('SELECT * FROM users');
}
