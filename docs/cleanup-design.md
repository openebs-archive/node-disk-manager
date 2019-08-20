# Clean-up of released BlockDevices

NDM Operator performs the cleanup of the data on the blockdevice(BD) before it can be claimed by 
another BlockDeviceClaim(BDC).

The cleanup operation can be of two types depending on the VolumeMode of the BDC. VolumeMode can be
- Block : a `wipefs` command will be issued on the BD
- FileSystem : an `rm -rf` command is issued on the mountpoint of the BD

The cleanup is performed by a kubernetes job, that is scheduled to run on a specified node. The
following cycle of operations is performed for scheduling a cleanup job.

```
+-----------------+         +------------------+         +------------------+         +---------------+             +------------+
|                 |         |  Check if status |         |  Check if job is |         | Check status  | Completed   |            |
| Reconcile loop  |         |    is released   |   Yes   |    in progress   |   No    |     of job    +-------------+ Delete job |
|    of BDx       +-------->+                  +-------->+                  +-------->+               | Successfully|            |
|                 |         |                  |         |                  |         |               |             |            |
+-----------------+         +------------------+         +--------+---------+         +-------+-------+             +------+-----+
                                                                  |                           |                            |	
                                                                  | Yes                       |Not Found                   |	
                                                                  v                           v                            |	
                                                          +-------+---------+          +------+-------+                    |	
                             +-----------------+          |                 |          |              |                    |	
                             |                 |    No    |  Check if BD is |          |   Check if   |                    |	
                             | Cancel the job  +<---------+  in active state|          | BD is active |                    |	
                             |                 |          |                 |          |              |                    |	
                             |                 |          |                 |          |              |                    |	
                             +-------+---------+          +-----------------+          +-------+------+                    |	
                                     |                             |                           |Yes                    	   |	
                                     |                             |                    +------+-----+                     |	
                                     |                             |                    |            |                     |	
                                     |                             | Yes                | start job  |                     |	
                                     v                             |                    |            |                     |	
                              +------+---------+                   |                    |            |                     |	
                              |                |                   |                    +-------+----+                     |	
                              |    return      |<------------------+----------------------------|--------------------------+ 	
                              +----------------+                                    		  			 
                                                                                                            	                 
```
