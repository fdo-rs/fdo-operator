package controllers

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	gomock "go.uber.org/mock/gomock"

	util "github.com/redhat-cop/operator-utils/pkg/util"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	fdov1alpha1 "github.com/empovit/fdo-operator/api/v1alpha1"
	"github.com/empovit/fdo-operator/internal/client"
)

var _ = Describe("FDOOnboardingServerReconciler", func() {
	Describe("Reconcile", func() {
		const (
			serverName = "test-server"
		)

		var (
			gCtrl *gomock.Controller
			c     *client.MockClient
			ctx   context.Context
			r     *FDOOnboardingServerReconciler
		)

		req := reconcile.Request{
			NamespacedName: types.NamespacedName{Name: serverName},
		}

		BeforeEach(func() {
			gCtrl = gomock.NewController(GinkgoT())
			c = client.NewMockClient(gCtrl)
			ctx = context.TODO()
		})

		When("a client error other than not-found occurs", func() {
			BeforeEach(func() {
				s := scheme.Scheme
				Expect(fdov1alpha1.AddToScheme(scheme.Scheme)).ToNot(HaveOccurred())

				r = &FDOOnboardingServerReconciler{
					ReconcilerBase: util.NewReconcilerBase(c, s, nil, nil, nil),
				}

				gomock.InOrder(
					c.EXPECT().
						Get(ctx, req.NamespacedName, gomock.Any()).
						Return(errors.New("generic error")),
				)
			})

			It("should not explicitly requeue and return an error", func() {
				res, err := r.Reconcile(ctx, req)
				Expect(err).To(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())
			})
		})
	})
})
