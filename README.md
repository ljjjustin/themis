## 项目简介

本项目的目标是让OpenStack集群的计算节点宕机后，可以自动的完成虚拟机的疏散，而不需要运维人员介入。

### 需求概述

[OpenStack](https://www.openstack.org/)是一套可用于构建私有云和公有云的一整套软件，但是OpenStack并不是一套成熟的商业产品。在使用OpenStack构建私有云或公有云的过程中，有大量的工作要做，平台的高可用就是其中的一项。随着集群规模的扩大，硬件设备发生故障的概率会逐步提升，如何避免单点故障以及如何降低硬件故障对云平台的影响，成为了云服务提供商必须面对和解决的问题。

在一个OpenStack集群中，提供虚拟机的计算节点通常是数量最多的服务器，因此发生故障的概率也最高，而云平台用户的业务就是部署在虚拟机里面。一旦计算节点出现硬件故障导致意外宕机，势必会影响其上运行的虚拟机，从而影响客户的业务。 在计算节点发生宕机时，能够快速的在其他计算节点重建并启动虚拟机是一个普遍的需求，一方面通过底层的分布式存储保证虚拟机的数据不丢失，另一方面通过快速的重建虚拟机，实现快速的业务恢复。

OpenStack对这个需求提供了基本的支持，OpenStack中负责计算资源创建和调度的子项目[Nova](https://docs.openstack.org/nova/latest/)提供了虚拟机[疏散接口](https://docs.openstack.org/nova/rocky/admin/evacuate.html)，但是OpenStack并没有提供整套完整的方案，也没有触及宕机后自动化疏散在实现上的难点：

  * 如何准确的判断服务器只是暂时性的网络不通，还是宕机了
  * 如何有效的隔离出现故障的服务器，保证故障后，集群能够快速的收敛到一个稳定可用的状态

本项目参考了[CloudStack](https://cloudstack.apache.org/)的[宿主机HA实现方案](https://cwiki.apache.org/confluence/display/CLOUDSTACK/Host+HA)[海云捷迅](https://www.awcloud.com/)在这方面的[实践](https://www.infoq.cn/article/OpenStack-awcloud-HA)，给出了可以实际落地的开源实现。

### 方案简介

### 安装部署

### 命令行工具

