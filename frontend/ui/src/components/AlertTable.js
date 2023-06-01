import React from 'react';
import { Card, Grid, Paper, Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@mui/material';

const alerts = [
  {
    name: 'CPU usage high',
    severity: 'critical',
    description: 'CPU usage is above 90%',
    firing: true,
  },
  {
    name: 'Memory usage high',
    severity: 'warning',
    description: 'Memory usage is above 80%',
    firing: true,
  },
  {
    name: 'Disk space low',
    severity: 'warning',
    description: 'Free disk space is below 10%',
    firing: true,
  },
  {
    name: 'HTTP error rate',
    severity: 'warning',
    description: 'HTTP error rate is above 5%',
    firing: true,
  },
  {
    name: 'Number of requests',
    severity: 'info',
    description: 'Number of requests is above 1000 per minute',
    firing: true,
  },
];

const AlertTable = () => {
  return (
      <Card sx={{ margin: "3rem auto" }}>
    <Grid container spacing={2}>
      <Grid item xs={12}>
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>Severity</TableCell>
                <TableCell>Description</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {alerts.map((alert, index) => (
                <TableRow key={index}>
                  <TableCell component="th" scope="row" sx={alert.firing ? { color: '#D32F2F' } : {}}>
                    {alert.name}
                  </TableCell>
                  <TableCell sx={alert.firing ? { color: '#D32F2F' } : {}}>{alert.severity}</TableCell>
                  <TableCell sx={alert.firing ? { color: '#D32F2F' } : {}}>{alert.description}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      </Grid>
    </Grid>
      </Card>
  );
};

export default AlertTable;
