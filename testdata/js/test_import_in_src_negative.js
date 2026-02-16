// Negative: Normal imports in source file
import express from 'express';
import { useState, useEffect } from 'react';
const lodash = require('lodash');

export function createApp() {
  return express();
}
